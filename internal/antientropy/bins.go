package antientropy

import (
	"fmt"
	"time"

	pb "github.com/drakeor/jmgr/api"
	"github.com/drakeor/jmgr/internal/store"
	"github.com/drakeor/jmgr/internal/utils"
	"github.com/willf/bloom"
)

// TODO: This is all pretty badly written and needs a refactor.
// I'm also not convinced Bloom filters are the best way to do this.
// I thought it would just be fun to try it out since I've always seem them on network switches.

// One bin is a chunk of time for a given bucket (e.g., 5m window)
type BinID string // e.g., "20250602T1130"

// A single bin's summary—Bloom is for all points in that bin
type BinSummary struct {
	Bin   BinID
	Bloom *bloom.BloomFilter
	Count int // for diagnostics only
}

// Per-bucket summary, listing bins present and their Bloom filters
type BucketSummary struct {
	Bucket string // e.g., "node1|cpu_pct|mount=/"
	Bins   []BinSummary
}

type NodeSummary map[string]*BucketSummary // map[bucket]BucketSummary

// Parameters for binning
const BinInterval = 5 * time.Minute // change here for new bin size

// BinForTimestamp returns the BinID for a given timestamp in milliseconds
func BinForTimestamp(ts int64) BinID {
	sec := ts / 1000
	t := time.Unix(sec, 0).UTC()
	binStart := t.Truncate(BinInterval)
	return BinID(binStart.Format("20060102T1504"))
}

// NewBloomSummary creates a Bloom filter from a slice of metric points
func NewBloomSummary(binID BinID, points []*pb.MetricPoint) BinSummary {
	// These parameters control Bloom accuracy: tweak as needed
	n := uint(len(points))
	if n == 0 {
		n = 1
	}

	// Allow for a 1% false positive rate
	filter := bloom.NewWithEstimates(n, 0.01)
	for _, pt := range points {
		b := []byte(fmt.Sprintf("%d:%.6f", pt.TsMs, pt.Value))
		filter.Add(b)
	}
	return BinSummary{Bin: binID, Bloom: filter, Count: len(points)}
}

// SummarizeStore extracts NodeSummary for all buckets/bins in the store
func SummarizeStore(s store.Driver) NodeSummary {
	out := make(NodeSummary)
	for _, bucket := range s.ListBuckets() {
		nodeID, name, labels, err := utils.ParseBucketKey(bucket)
		if err != nil {
			continue // skip broken bucket keys
		}
		pts, _ := s.Get(nodeID, name, time.Unix(0, 0), labels)

		binMap := map[BinID][]*pb.MetricPoint{}
		for _, pt := range pts {
			bin := BinForTimestamp(pt.TsMs)
			binMap[bin] = append(binMap[bin], pt)
		}
		summ := &BucketSummary{Bucket: bucket}
		for binID, points := range binMap {
			summ.Bins = append(summ.Bins, NewBloomSummary(binID, points))
		}
		out[bucket] = summ
	}
	return out
}

func ToProtoNodeSummary(ns NodeSummary) *pb.NodeSummary {
	var out pb.NodeSummary
	for _, bs := range ns {
		var pbBins []*pb.BinSummary
		for _, bin := range bs.Bins {
			bloomBytes, _ := bin.Bloom.GobEncode()
			pbBins = append(pbBins, &pb.BinSummary{
				BinId:      string(bin.Bin),
				Count:      uint32(bin.Count),
				BloomBytes: bloomBytes,
			})
		}
		out.Summaries = append(out.Summaries, &pb.BucketSummary{
			Bucket: bs.Bucket,
			Bins:   pbBins,
		})
	}
	return &out
}

func FromProtoNodeSummary(ps *pb.NodeSummary) NodeSummary {
	ns := make(NodeSummary)
	for _, bs := range ps.Summaries {
		var bins []BinSummary
		for _, bin := range bs.Bins {
			bf := bloom.New(1, 1) // size doesn't matter, will reset below
			bf.GobDecode(bin.BloomBytes)
			bins = append(bins, BinSummary{
				Bin:   BinID(bin.BinId),
				Bloom: bf,
				Count: int(bin.Count),
			})
		}
		ns[bs.Bucket] = &BucketSummary{
			Bucket: bs.Bucket,
			Bins:   bins,
		}
	}
	return ns
}
