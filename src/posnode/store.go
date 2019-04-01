package posnode

import (
	"time"

	"github.com/dgraph-io/badger"
	"github.com/golang/protobuf/proto"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/kvdb"
	"github.com/Fantom-foundation/go-lachesis/src/posnode/api"
)

// Store is a node persistent storage working over physical key-value database.
type Store struct {
	physicalDB kvdb.Database

	peers       kvdb.Database
	discovery   kvdb.Database
	topPeers    kvdb.Database
	knownPeers  kvdb.Database
	peerHeights kvdb.Database
}

// NewMemStore creates store over memory map.
func NewMemStore() *Store {
	s := &Store{
		physicalDB: kvdb.NewMemDatabase(),
	}
	s.init()
	return s
}

// NewBadgerStore creates store over badger database.
func NewBadgerStore(db *badger.DB) *Store {
	s := &Store{
		physicalDB: kvdb.NewBadgerDatabase(db),
	}
	s.init()
	return s
}

func (s *Store) init() {
	s.peers = kvdb.NewTable(s.physicalDB, "peer_")
	s.topPeers = kvdb.NewTable(s.physicalDB, "top_peers_")
	s.knownPeers = kvdb.NewTable(s.physicalDB, "known_peers_")
	s.peerHeights = kvdb.NewTable(s.physicalDB, "peer_height_")
	s.discovery = kvdb.NewTable(s.physicalDB, "discovery_")
}

// Close leaves underlying database.
func (s *Store) Close() {
	s.peerHeights = nil
	s.knownPeers = nil
	s.topPeers = nil
	s.peers = nil
	s.discovery = nil
	s.physicalDB.Close()
}

// SetPeer stores peer.
func (s *Store) SetPeer(peer *Peer) {
	w := peer.ToWire()
	s.set(s.peers, peer.ID.Bytes(), w)
}

// GetPeerInfo returns stored peer info.
// Result is a ready gRPC message.
func (s *Store) GetPeerInfo(id hash.Peer) *api.PeerInfo {
	w, _ := s.get(s.peers, id.Bytes(), &api.PeerInfo{}).(*api.PeerInfo)
	return w
}

// GetPeer returns stored peer.
func (s *Store) GetPeer(id hash.Peer) *Peer {
	w := s.GetPeerInfo(id)
	if w == nil {
		return nil
	}

	return WireToPeer(w)
}

// BootstrapPeers stores peer list.
func (s *Store) BootstrapPeers(peers []*Peer) {
	if len(peers) < 1 {
		return
	}

	// save peers
	batch := s.peers.NewBatch()
	defer batch.Reset()

	ids := make([]hash.Peer, len(peers))
	for i, peer := range peers {
		var pbf proto.Buffer
		w := peer.ToWire()
		if err := pbf.Marshal(w); err != nil {
			panic(err)
		}
		if err := batch.Put(peer.ID.Bytes(), pbf.Bytes()); err != nil {
			panic(err)
		}
		ids[i] = peer.ID
	}

	if err := batch.Write(); err != nil {
		panic(err)
	}

	// add peers to top
	s.SetTopPeers(ids)
}

// SetTopPeers stores peers.top.
func (s *Store) SetTopPeers(ids []hash.Peer) {
	var key = []byte("current")
	w := IDsToWire(ids)
	s.set(s.topPeers, key, w)
}

// GetTopPeers returns peers.top.
func (s *Store) GetTopPeers() []hash.Peer {
	var key = []byte("current")
	w, _ := s.get(s.topPeers, key, &api.PeersID{}).(*api.PeersID)
	return WireToIDs(w)
}

// SetKnownPeers stores all peers ID.
func (s *Store) SetKnownPeers(ids []hash.Peer) {
	var key = []byte("current")
	w := IDsToWire(ids)
	s.set(s.knownPeers, key, w)
}

// GetKnownPeers returns all peers ID.
func (s *Store) GetKnownPeers() []hash.Peer {
	var key = []byte("current")
	w, _ := s.get(s.knownPeers, key, &api.PeersID{}).(*api.PeersID)
	return WireToIDs(w)
}

// SetPeerHeight stores last event index of peer.
func (s *Store) SetPeerHeight(id hash.Peer, height uint64) {
	if err := s.peerHeights.Put(id.Bytes(), intToBytes(height)); err != nil {
		panic(err)
	}
}

// GetPeerHeight returns last event index of peer.
func (s *Store) GetPeerHeight(id hash.Peer) uint64 {
	buf, err := s.peerHeights.Get(id.Bytes())
	if err != nil {
		panic(err)
	}
	if buf == nil {
		return 0
	}

	return bytesToInt(buf)
}

// GetDiscoveryInfo returns stored discovery info.
func (s *Store) GetDiscoveryInfo(id hash.Peer) *api.DiscoveryInfo {
	w, _ := s.get(s.discovery, id.Bytes(), &api.DiscoveryInfo{}).(*api.DiscoveryInfo)
	return w
}

// GetDiscovery returns stored discovery.
func (s *Store) GetDiscovery(id hash.Peer) *Discovery {
	w := s.GetDiscoveryInfo(id)
	if w == nil {
		return nil
	}

	return WireToDiscovery(w)
}

// SetDiscovery returns stored discovery info.
func (s *Store) SetDiscovery(discovery *Discovery) {
	w := discovery.ToWire()
	s.set(s.discovery, discovery.ID.Bytes(), w)
}

// SetDiscoveryAvailability sets availability and
// last request time.
func (s *Store) SetDiscoveryAvailability(d *Discovery, available bool) {
	d.Available = available
	d.LastRequest = time.Now()
	s.SetDiscovery(d)
}

/*
 * Utils:
 */

func (s *Store) set(table kvdb.Database, key []byte, val proto.Message) {
	var pbf proto.Buffer

	if err := pbf.Marshal(val); err != nil {
		panic(err)
	}

	if err := table.Put(key, pbf.Bytes()); err != nil {
		panic(err)
	}
}

func (s *Store) get(table kvdb.Database, key []byte, to proto.Message) proto.Message {
	buf, err := table.Get(key)
	if err != nil {
		panic(err)
	}
	if buf == nil {
		return nil
	}

	err = proto.Unmarshal(buf, to)
	if err != nil {
		panic(err)
	}
	return to
}

func (s *Store) has(table kvdb.Database, key []byte) bool {
	res, err := table.Has(key)
	if err != nil {
		panic(err)
	}
	return res
}

func intToBytes(n uint64) []byte {
	var res [8]byte
	for i := 0; i < len(res); i++ {
		res[i] = byte(n)
		n = n >> 8
	}
	return res[:]
}

func bytesToInt(b []byte) uint64 {
	var res uint64
	for i := 0; i < len(b); i++ {
		res += uint64(b[i]) << uint(i*8)
	}
	return res
}
