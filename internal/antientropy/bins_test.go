package antientropy

import (
	"fmt"
	"testing"
	"time"

	pb "github.com/drakeor/jmgr/api"
	"github.com/willf/bloom"
)

func TestBinForTimestamp(t *testing.T) {
	// Should consistently bucket same time to same bin
	ts := time.Date(2025, 6, 1, 12, 7, 0, 0, time.UTC).UnixMilli()
	bin := BinForTimestamp(ts)
	expected := BinID("20250601T1205") // if BinInterval = 5m
	if bin != expected {
		t.Fatalf("got %q want %q", bin, expected)
	}
}
func TestNewBloomSummaryAndSerialization(t *testing.T) {
	points := []*pb.MetricPoint{
		{TsMs: 1000, Value: 1},
		{TsMs: 2000, Value: 2},
	}
	bin := NewBloomSummary("20250601T1200", points)
	// Should test positive for added points
	for _, pt := range points {
		b := []byte(fmt.Sprintf("%d:%.6f", pt.TsMs, pt.Value))
		if !bin.Bloom.Test(b) {
			t.Errorf("expected bloom to contain %v", b)
		}
	}

	// Test proto encode/decode
	data, err := bin.Bloom.GobEncode()
	if err != nil {
		t.Fatalf("GobEncode failed: %v", err)
	}
	bf := bloom.New(bin.Bloom.Cap(), bin.Bloom.K())
	if err := bf.GobDecode(data); err != nil {
		t.Fatalf("GobDecode failed: %v", err)
	}
	b := []byte(fmt.Sprintf("%d:%.6f", points[0].TsMs, points[0].Value))
	if !bf.Test(b) {
		t.Error("deserialized bloom should contain point")
	}
}
