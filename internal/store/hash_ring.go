package store

import (
	"hash/crc32"
	"sort"
	"sync"
)

// hashRing provides k‑successor lookups on a static set of node IDs.
// It is concurrency‑safe.
type hashRing struct {
	hashes []uint32 // sorted hashes
	nodes  []string // nodeID for each hash
	mu     sync.RWMutex
}

// newHashRing creates a new hash ring with the given peer IDs.
func newHashRing(ids []string) *hashRing {
	r := &hashRing{}
	r.Rebuild(ids)
	return r
}

// Rebuild replaces the ring with a new set of peer IDs.
func (r *hashRing) Rebuild(ids []string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	type entry struct {
		hash uint32
		id   string
	}
	var entries []entry
	for _, id := range ids {
		h := crc32.ChecksumIEEE([]byte(id))
		entries = append(entries, entry{hash: h, id: id})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].hash < entries[j].hash
	})
	r.hashes = make([]uint32, len(entries))
	r.nodes = make([]string, len(entries))
	for i, e := range entries {
		r.hashes[i] = e.hash
		r.nodes[i] = e.id
	}
}

// successors returns the next k unique node IDs clockwise from hash h.
func (r *hashRing) successors(hash uint32, k int) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	n := len(r.hashes)
	if n == 0 {
		return nil
	}
	// Find first hash >= h
	start := 0
	for i, hval := range r.hashes {
		if hash <= hval {
			start = i
			break
		}
	}
	var res []string
	seen := map[string]struct{}{}
	i := start
	for len(res) < k && i-start < n*2 { // wrap around at most twice
		nodeID := r.nodes[i%n]
		if _, ok := seen[nodeID]; !ok {
			res = append(res, nodeID)
			seen[nodeID] = struct{}{}
		}
		i++
	}
	return res
}
