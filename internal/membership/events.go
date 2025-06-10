package membership

import pb "github.com/drakeor/jmgr/api"

type EventType int

const (
	EventJoin EventType = iota
	EventLeave
)

type Event struct {
	Type EventType
	Peer *pb.PeerInfo
}

// verbose hook: how many peers were merged from which source digest
type MergeEvent struct {
	Source string
	Count  int
}
