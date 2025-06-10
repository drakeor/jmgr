package antientropy

// SyncJob represents a job to sync a specific bin from a peer.
type SyncJob struct {
	PeerID string
	Bucket string
	Bin    BinID
}
