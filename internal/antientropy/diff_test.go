package antientropy

import (
	"testing"

	pb "github.com/drakeor/jmgr/api"
)

func TestDiffSummaries(t *testing.T) {
	local := NodeSummary{
		"bucketA": &BucketSummary{
			Bucket: "bucketA",
			Bins: []BinSummary{
				NewBloomSummary("bin1", []*pb.MetricPoint{{TsMs: 1, Value: 10}}),
			},
		},
	}
	remote := NodeSummary{
		"bucketA": &BucketSummary{
			Bucket: "bucketA",
			Bins: []BinSummary{
				NewBloomSummary("bin1", []*pb.MetricPoint{{TsMs: 1, Value: 10}}),
				NewBloomSummary("bin2", []*pb.MetricPoint{{TsMs: 2, Value: 20}}),
			},
		},
	}
	jobs := DiffSummaries(local, remote, "peer1")
	if len(jobs) != 1 || jobs[0].Bin != "bin2" {
		t.Errorf("expected 1 job for bin2, got %+v", jobs)
	}
}
