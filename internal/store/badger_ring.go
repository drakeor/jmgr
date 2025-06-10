package store

import (
	"fmt"
	"hash/crc32"
	"strings"
	"time"

	badger "github.com/dgraph-io/badger/v4"
	pb "github.com/drakeor/jmgr/api"
	"github.com/drakeor/jmgr/internal/utils"
	"google.golang.org/protobuf/proto"
)

type BadgerRing struct {
	selfID   string
	db       *badger.DB
	ring     *hashRing
	maxBytes int64
	meta     *Meta
}

// BadgerRing implements the Driver interface using BadgerDB.
var _ Driver = (*BadgerRing)(nil)

// NewBadgerRing creates a new BadgerRing instance with the given parameters.
// selfID is the ID of this node, peerIDs are the IDs of other nodes in the ring,
func NewBadgerRing(selfID string, peerIDs []string, dir string, maxBytes int64, inMem bool) (*BadgerRing, error) {
	opts := badger.DefaultOptions(dir)
	if inMem {
		opts = opts.WithInMemory(true)
	}
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	meta, err := NewMeta(dir)
	if err != nil {
		db.Close()
		return nil, err
	}
	return &BadgerRing{
		selfID:   selfID,
		db:       db,
		ring:     newHashRing(append(peerIDs, selfID)),
		maxBytes: maxBytes,
		meta:     meta,
	}, nil
}

// makeKey builds the metric key for storage and retrieval.
// If ts == 0, return the prefix (for Get); if ts > 0, return full key (for Put).
func makeKey(nodeID, name string, labels map[string]string, ts int64) []byte {
	labelKey := utils.SerializeLabels(labels)
	if ts == 0 {
		return []byte(fmt.Sprintf("%s|%s|%s|", nodeID, name, labelKey))
	}
	return []byte(fmt.Sprintf("%s|%s|%s|%d", nodeID, name, labelKey, ts))
}

// Put writes the metric if this node is one of the two replicas.
func (b *BadgerRing) Put(m *pb.Metric) error {
	var k int

	// Translate tier to number of replicas
	switch m.Tier {
	case "hot":
		k = len(b.ring.nodes) // all nodes
	case "warm":
		k = 3
	case "cold":
		k = 2
	default:
		panic(fmt.Sprintf("unknown tier: %q", m.Tier))
	}

	// Get the successors for this metric's node ID
	h := crc32.ChecksumIEEE([]byte(m.NodeId))
	reps := b.ring.successors(h, k)
	storeHere := false
	for _, id := range reps {
		if id == b.selfID {
			storeHere = true
			break
		}
	}
	//fmt.Printf("Metric for %s assigned to: %v (storeHere=%v)\n", m.NodeId, reps, storeHere)

	if !storeHere {
		return nil
	}
	return b.db.Update(func(txn *badger.Txn) error {
		for _, p := range m.Points {
			key := makeKey(m.NodeId, m.Name, m.Labels, p.TsMs)
			val, _ := proto.Marshal(p)
			if err := txn.Set(key, val); err != nil {
				return err
			}
		}
		return nil
	})
}

// Get retrieves all metric points for the given nodeID, name, and labels since the specified time.
// It returns all points with timestamps greater than or equal to 'since'.
// If no points are found, it returns an empty slice.
func (b *BadgerRing) Get(nodeID, name string, since time.Time, labels map[string]string) ([]*pb.MetricPoint, error) {
	var out []*pb.MetricPoint
	prefix := makeKey(nodeID, name, labels, 0) // all keys for nodeID+name+labels
	err := b.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			val, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			var p pb.MetricPoint
			if err := proto.Unmarshal(val, &p); err != nil {
				return err
			}
			if p.TsMs >= since.UnixMilli() {
				out = append(out, &p)
			}
		}
		return nil
	})
	return out, err
}

// Vacuum removes old entries until the disk usage is below maxBytes.
func (b *BadgerRing) Vacuum(maxBytes int64) error {
	if maxBytes == 0 {
		maxBytes = b.maxBytes
	}
	if maxBytes == 0 {
		return nil
	}
	if b.DiskUsage() <= maxBytes {
		return nil
	}
	return b.db.Update(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.IteratorOptions{PrefetchValues: false})
		defer it.Close()
		it.Rewind()
		for ; it.Valid() && b.DiskUsage() > maxBytes; it.Next() {
			if err := txn.Delete(it.Item().Key()); err != nil {
				return err
			}
		}
		return nil
	})
}

// DiskUsage returns the total disk usage in bytes, including live data and tables.
func (b *BadgerRing) DiskUsage() int64 {
	live, tables := b.db.Size()
	return live + tables
}

// Close closes the BadgerDB instance.
func (b *BadgerRing) Close() error {
	return b.db.Close()
}

// NextIncarnation increments the incarnation number and returns the new value.
func (b *BadgerRing) NextIncarnation() (uint64, error) {
	return b.meta.NextIncarnation()
}

// GetIncarnation retrieves the current incarnation number.
func (b *BadgerRing) GetIncarnation() uint64 {
	return b.meta.GetIncarnation()
}

// ListBuckets returns a list of unique buckets in the store.
func (b *BadgerRing) ListBuckets() []string {
	buckets := make(map[string]struct{})
	b.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			k := string(it.Item().Key())
			// k is nodeID|name|labels|ts
			parts := strings.SplitN(k, "|", 4)
			if len(parts) >= 3 {
				bucket := fmt.Sprintf("%s|%s|%s", parts[0], parts[1], parts[2])
				buckets[bucket] = struct{}{}
			}
		}
		return nil
	})
	out := make([]string, 0, len(buckets))
	for k := range buckets {
		out = append(out, k)
	}
	return out
}
