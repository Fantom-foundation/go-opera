package posnode

import (
	"github.com/golang/protobuf/proto"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/wire"
	"github.com/Fantom-foundation/go-lachesis/src/kvdb"
	"github.com/Fantom-foundation/go-lachesis/src/logger"
	"github.com/Fantom-foundation/go-lachesis/src/posnode/api"
)

// Store is a node persistent storage working over physical key-value database.
type Store struct {
	physicalDB kvdb.Database

	peers       kvdb.Database
	peersTop    kvdb.Database
	peerHeights kvdb.Database

	events kvdb.Database
	hashes kvdb.Database

	logger.Instance
}

// NewStore creates store over key-value db.
func NewStore(db kvdb.Database) *Store {
	s := &Store{
		physicalDB: db,
		Instance:   logger.MakeInstance(),
	}
	s.init()
	return s
}

// NewMemStore creates store over memory map.
func NewMemStore() *Store {
	db := kvdb.NewMemDatabase()
	return NewStore(db)
}

func (s *Store) init() {
	s.peers = kvdb.NewTable(s.physicalDB, "peer_")
	s.peersTop = kvdb.NewTable(s.physicalDB, "top_peers_")
	s.peerHeights = kvdb.NewTable(s.physicalDB, "peer_height_")

	s.events = kvdb.NewTable(s.physicalDB, "event_")
	s.hashes = kvdb.NewTable(s.physicalDB, "hash_")
}

// Close leaves underlying database.
func (s *Store) Close() {
	s.peerHeights = nil
	s.peersTop = nil
	s.peers = nil
	s.events = nil
	s.hashes = nil
	s.physicalDB.Close()
}

// SetEvent stores event.
func (s *Store) SetEvent(e *inter.Event) {
	s.set(s.events, e.Hash().Bytes(), e.ToWire())
}

// GetWireEvent returns stored event.
// Result is a ready gRPC message.
func (s *Store) GetWireEvent(h hash.Event) *wire.Event {
	w, _ := s.get(s.events, h.Bytes(), &wire.Event{}).(*wire.Event)
	return w
}

// GetEvent returns stored event.
func (s *Store) GetEvent(h hash.Event) *inter.Event {
	w := s.GetWireEvent(h)
	return inter.WireToEvent(w)
}

// SetEventHash stores hash.
func (s *Store) SetEventHash(creator hash.Peer, index uint64, hash hash.Event) {
	key := append(creator.Bytes(), intToBytes(index)...)

	if err := s.hashes.Put(key, hash.Bytes()); err != nil {
		s.Fatal(err)
	}
}

// GetEventHash returns stored event hash.
func (s *Store) GetEventHash(creator hash.Peer, index uint64) *hash.Event {
	key := append(creator.Bytes(), intToBytes(index)...)

	buf, err := s.hashes.Get(key)
	if err != nil {
		s.Fatal(err)
	}
	if buf == nil {
		return nil
	}

	e := hash.BytesToEventHash(buf)
	return &e
}

// HasEvent returns true if event exists.
func (s *Store) HasEvent(h hash.Event) bool {
	return s.has(s.events, h.Bytes())
}

// SetWirePeer stores peer info.
func (s *Store) SetWirePeer(id hash.Peer, info *api.PeerInfo) {
	s.set(s.peers, id.Bytes(), info)
}

// SetPeer stores peer.
func (s *Store) SetPeer(peer *Peer) {
	info := peer.ToWire()
	s.SetWirePeer(peer.ID, info)
}

// GetWirePeer returns stored peer info.
// Result is a ready gRPC message.
func (s *Store) GetWirePeer(id hash.Peer) *api.PeerInfo {
	w, _ := s.get(s.peers, id.Bytes(), &api.PeerInfo{}).(*api.PeerInfo)
	return w
}

// GetPeer returns stored peer.
func (s *Store) GetPeer(id hash.Peer) *Peer {
	w := s.GetWirePeer(id)
	return WireToPeer(w)
}

// BootstrapPeers stores peer list.
func (s *Store) BootstrapPeers(peers ...*Peer) {
	if len(peers) < 1 {
		return
	}

	// save peers
	batch := s.peers.NewBatch()
	defer batch.Reset()

	ids := make([]hash.Peer, 0, len(peers))
	for _, peer := range peers {
		// skip empty
		if peer == nil || peer.PubKey == nil || peer.ID.IsEmpty() || peer.Host == "" {
			continue
		}

		var pbf proto.Buffer
		w := peer.ToWire()
		if err := pbf.Marshal(w); err != nil {
			s.Fatal(err)
		}
		if err := batch.Put(peer.ID.Bytes(), pbf.Bytes()); err != nil {
			s.Fatal(err)
		}
		ids = append(ids, peer.ID)
	}

	if err := batch.Write(); err != nil {
		s.Fatal(err)
	}

	s.SetTopPeers(ids)
}

// SetTopPeers stores peers.top.
func (s *Store) SetTopPeers(ids []hash.Peer) {
	var key = []byte("current")
	w := IDsToWire(ids)
	s.set(s.peersTop, key, w)
}

// GetTopPeers returns peers.top.
func (s *Store) GetTopPeers() []hash.Peer {
	var key = []byte("current")
	w, _ := s.get(s.peersTop, key, &api.PeerIDs{}).(*api.PeerIDs)
	return WireToIDs(w)
}

// SetPeerHeight stores last event index of peer.
func (s *Store) SetPeerHeight(id hash.Peer, height uint64) {
	if err := s.peerHeights.Put(id.Bytes(), intToBytes(height)); err != nil {
		s.Fatal(err)
	}
}

// GetPeerHeight returns last event index of peer.
func (s *Store) GetPeerHeight(id hash.Peer) uint64 {
	buf, err := s.peerHeights.Get(id.Bytes())
	if err != nil {
		s.Fatal(err)
	}
	if buf == nil {
		return 0
	}

	return bytesToInt(buf)
}

/*
 * Utils:
 */

func (s *Store) set(table kvdb.Database, key []byte, val proto.Message) {
	var pbf proto.Buffer

	if err := pbf.Marshal(val); err != nil {
		s.Fatal(err)
	}

	if err := table.Put(key, pbf.Bytes()); err != nil {
		s.Fatal(err)
	}
}

func (s *Store) get(table kvdb.Database, key []byte, to proto.Message) proto.Message {
	buf, err := table.Get(key)
	if err != nil {
		s.Fatal(err)
	}
	if buf == nil {
		return nil
	}

	err = proto.Unmarshal(buf, to)
	if err != nil {
		s.Fatal(err)
	}
	return to
}

func (s *Store) has(table kvdb.Database, key []byte) bool {
	res, err := table.Has(key)
	if err != nil {
		s.Fatal(err)
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
