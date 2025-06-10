package connections

// PeerMeta contains metadata about a peer in the cluster.
type PeerMeta struct {
	Addr        string
	Incarnation uint64
}
