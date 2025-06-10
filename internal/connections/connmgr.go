package connections

import (
	"context"
	"log"
	"regexp"
	"strings"
	"sync"

	pb "github.com/drakeor/jmgr/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
)

// Dead code for now
// Used for traffic classification for a QoS idea I had that's half-baked.
type TrafficClass int

const (
	TrafficHeartbeat TrafficClass = iota
	TrafficData
	TrafficMetrics
	TrafficSync
)

// Holds state for the connection manager
type ConnManager struct {
	mu      sync.Mutex
	conns   map[string]*grpc.ClientConn
	dialing map[string]*dialState
	peers   map[string]*PeerMeta // peerID -> meta (addr, incarnation)
}

// Holds dial state for a peer
type dialState struct {
	ready chan struct{}
}

// Creates a new connection manager instance
func NewConnManager() *ConnManager {
	return &ConnManager{
		conns:   make(map[string]*grpc.ClientConn),
		dialing: make(map[string]*dialState),
		peers:   make(map[string]*PeerMeta),
	}
}

// Helper to check if a string looks like an address (IP:port)
// TODO: Move to utils
func looksLikeAddr(s string) bool {
	return strings.Count(s, ":") == 1 && strings.ContainsAny(s, "0123456789")
}

// Helper function to validate peerID format
// TODO: Move to utils
// Valid peerID: alphanumeric, dashes, underscores, colons, dots, and square brackets
var validPeerID = regexp.MustCompile(`^[A-Za-z0-9\-_:\.\[\]]+$`)

func isValidPeerID(s string) bool {
	return validPeerID.MatchString(s)
}

// Add or update peer
// Note that sometimes we just only know a peer's address, which is a valid PeerID.
// However, when we learn the canonical nodeID for that address, we should
// prioritize that instead as the primary ID.
// Thus, this function dedupes address-as-ID in favor of nodeID
func (m *ConnManager) AddPeer(peerID, addr string) {

	// We do not allow instances if peerID AND addr are BOTH empty!
	// (This screwed me over)
	if peerID == "" || addr == "" {
		log.Printf("[connmgr] ERROR: ignoring incomplete peer (peerID='%s', addr='%s')", peerID, addr)
		return
	}
	if !isValidPeerID(peerID) {
		log.Fatalf("[connmgr] FATAL: Invalid peerID: '%s'", peerID)
		panic("Invalid peerID in AddPeer")
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	// Only update/add if the peerID is missing or has a different address
	if meta, ok := m.peers[peerID]; ok && meta.Addr == addr {
		// Already exists, skip log and update
		//log.Printf("[connmgr] AddPeer: peer '%s' already exists with addr '%s', skipping add", peerID, addr)
		return
	}

	// Don't add address-as-ID if we already have a canonical peerID for this addr
	// Canonical ID has priority over address-as-ID
	if peerID == addr {
		for id, meta := range m.peers {
			if id != peerID && meta.Addr == addr && !looksLikeAddr(id) {
				log.Printf("[connmgr] AddPeer: ignoring address-as-ID '%s' because canonical peerID '%s' exists for addr '%s'", peerID, id, addr)
				return
			}
		}
	}

	// Remove any address-as-ID entry if this addr is now claimed by a canonical peerID
	// Canonical ID has priority over address-as-ID
	if !looksLikeAddr(peerID) {
		for k, meta := range m.peers {
			if k != peerID && looksLikeAddr(k) && meta.Addr == addr {
				log.Printf("[connmgr] removing address-as-ID '%s' now that canonical '%s' found", k, peerID)
				delete(m.peers, k)
				if conn, ok := m.conns[k]; ok {
					conn.Close()
					delete(m.conns, k)
				}
			}
		}
	}

	// Do the actual add/update
	log.Printf("[connmgr] AddPeer: setting peer '%s' -> '%s'", peerID, addr)
	meta, ok := m.peers[peerID]
	if !ok || meta.Addr != addr {
		m.peers[peerID] = &PeerMeta{Addr: addr}
	}
	// Print all peers for debugging
	/*log.Printf("[connmgr] Current peers:")
	for id, meta := range m.peers {
		log.Printf("  PeerID: '%s', Addr: '%s', Incarnation: %d", id, meta.Addr, meta.Incarnation)
	}*/
}

// Remove a peer (and close its connection if open)
func (m *ConnManager) RemovePeer(peerID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.peers, peerID)
	if conn, ok := m.conns[peerID]; ok {
		conn.Close()
		delete(m.conns, peerID)
	}
}

// Normalize peerID: if looks like an address, find real nodeID if available
// If we don't have a canonicalPeerID, fallback to using the address as peerID.
func (m *ConnManager) canonicalPeerID(peerID string) (string, string) {
	if !looksLikeAddr(peerID) {
		return peerID, ""
	}
	for id, meta := range m.peers {
		if meta.Addr == peerID && !looksLikeAddr(id) {
			return id, peerID
		}
	}
	return peerID, "" // fallback: address as peerID
}

// Returns a single shared *grpc.ClientConn for each peerID (deduped).
// This prevents having an unbounded number of connections to the same peer.
func (m *ConnManager) Get(ctx context.Context, peerID, addr string, class TrafficClass) (*grpc.ClientConn, error) {
	m.mu.Lock()
	// Get the canonical peerID and address
	realID, altAddr := m.canonicalPeerID(peerID)
	if altAddr != "" && addr == "" {
		addr = altAddr
	}

	// If we know the canonical realID, update mapping/address
	if addr != "" {
		meta, ok := m.peers[realID]
		if !ok || meta.Addr != addr {
			m.peers[realID] = &PeerMeta{Addr: addr}
		}
	}
	peerID = realID

	// If we have a connection for this peer, return it
	if conn, ok := m.conns[peerID]; ok && conn != nil && conn.GetState() != connectivity.Shutdown {
		m.mu.Unlock()
		return conn, nil
	}

	// If we are already dialing this peer, wait for it to be ready
	if state, waiting := m.dialing[peerID]; waiting {
		m.mu.Unlock()
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-state.ready:
		}
		return m.Get(ctx, peerID, addr, class)
	}
	state := &dialState{ready: make(chan struct{})}
	m.dialing[peerID] = state
	m.mu.Unlock()

	// If we reach here, we need to open a new connection to the peer
	conn, err := grpc.DialContext(ctx, addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	m.mu.Lock()
	delete(m.dialing, peerID)
	if err == nil {
		m.conns[peerID] = conn
	}
	close(state.ready)
	m.mu.Unlock()
	return conn, err
}

// Close a specific connection by peerID
func (m *ConnManager) Close(peerID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if conn, ok := m.conns[peerID]; ok {
		err := conn.Close()
		delete(m.conns, peerID)
		return err
	}
	return nil
}

// CloseAll closes all connections and clears the peer list
func (m *ConnManager) CloseAll() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for pid, conn := range m.conns {
		conn.Close()
		delete(m.conns, pid)
	}
	for pid := range m.peers {
		delete(m.peers, pid)
	}
}

// Return all peer IDs (excluding address-as-ID if we know realID for that addr)
// Also, purge any entry where BOTH id and address are empty! (Sanity check)
func (m *ConnManager) Peers() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]string, 0, len(m.peers))

	// Building an address to nodeID mapping for deduping
	addrToID := make(map[string]string)
	for id, meta := range m.peers {

		// Sanity check: if both ID and address are empty, remove this entry
		if id == "" && (meta == nil || meta.Addr == "") {
			log.Printf("[connmgr] removing broken peer entry: empty ID AND address")
			delete(m.peers, id)
			continue
		}

		// Ignore empty addresses
		if meta == nil || meta.Addr == "" {
			continue
		}

		// If this is address-as-ID, map it to the canonical nodeID if available
		if !looksLikeAddr(id) {
			addrToID[meta.Addr] = id
		}
	}

	// Build the actual list of peer IDs
	for id, meta := range m.peers {
		// If this is address-as-ID, skip if canonical nodeID exists for this address
		if looksLikeAddr(id) {
			if nodeID, ok := addrToID[meta.Addr]; ok && nodeID != id {
				continue // skip redundant address-as-ID
			}
		}
		out = append(out, id)
	}
	return out
}

// Get peer address by ID or address (always canonicalizes if possible)
func (m *ConnManager) AddrForPeer(peerID string) (string, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	realID, _ := m.canonicalPeerID(peerID)
	meta, ok := m.peers[realID]
	if !ok {
		return "", false
	}
	return meta.Addr, true
}

// Return all peer infos (for diagnostics/debug/UI)
// Also, purge any entry where BOTH id and address are empty!
func (m *ConnManager) PeerInfos() map[string]*PeerMeta {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make(map[string]*PeerMeta, len(m.peers))
	for k, v := range m.peers {
		if k == "" && (v == nil || v.Addr == "") {
			log.Printf("[connmgr] removing broken peer info: empty ID AND address")
			delete(m.peers, k)
			continue
		}
		out[k] = v
	}
	return out
}

// MergeDigest: update peer addresses from a MembershipDigest (deduping logic)
// Rejects digest entries where BOTH Id and Addr are empty!
func (m *ConnManager) MergeDigest(d *pb.MembershipDigest, selfID string) (updated []string) {

	// Sanity check for digest
	if d == nil {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	// Iterate over the peers in the digest
	for _, pin := range d.Peers {

		// Sanity check: if both ID and Addr are empty, skip this entry
		if pin.Id == "" && pin.Addr == "" {
			log.Printf("[connmgr] skipping digest entry with empty ID AND address")
			continue
		}
		if pin.Id == selfID || pin.Addr == "" {
			continue
		}
		// Remove any address-as-ID entry matching this addr but different key (deduplication)
		for k, meta := range m.peers {
			if k != pin.Id && looksLikeAddr(k) && meta.Addr == pin.Addr {
				log.Printf("[connmgr] removing address-as-ID '%s' (digest update) for canonical '%s'", k, pin.Id)
				delete(m.peers, k)
				if conn, ok := m.conns[k]; ok {
					conn.Close()
					delete(m.conns, k)
				}
			}
		}
		// Update or add the peer info
		meta, ok := m.peers[pin.Id]
		if !ok || meta.Addr != pin.Addr || meta.Incarnation != pin.Incarnation {
			log.Printf("[connmgr] MergeDigest: updating peer '%s' -> '%s'", pin.Id, pin.Addr)
			m.peers[pin.Id] = &PeerMeta{Addr: pin.Addr, Incarnation: pin.Incarnation}
			updated = append(updated, pin.Id)
		}
	}
	return
}
