package posnode

import (
	"github.com/dgraph-io/badger"
	"github.com/go-errors/errors"
	"github.com/golang/protobuf/proto"

	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/kvdb"
	"github.com/Fantom-foundation/go-lachesis/src/posnode/wire"
)

// Store is a node persistent storage working over physical key-value database.
type Store struct {
	physicalDB kvdb.Database

	peers        kvdb.Database
	top10PeersID kvdb.Database
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
	s.top10PeersID = kvdb.NewTable(s.physicalDB, "top10PeersID_")
}

// Close leaves underlying database.
func (s *Store) Close() {
	s.peers = nil
	s.top10PeersID = nil
	s.physicalDB.Close()
}

// SetTopPeersID stores peers ID.
func (s *Store) SetTopPeersID(ids []common.Address) error {
	length := len(ids)

	if length > 10 {
		return errors.New("Error: size of array more than 10")
	}

	addresses := make([]string, length)

	// TODO: too slow solution
	for _, id := range ids {
		addresses = append(addresses, id.Hex())
	}

	w := &wire.PeersID{
		ID: addresses,
	}

	return s.set(s.top10PeersID, []byte{0}, w)
}

// GetTopPeersID returns stored peer.
func (s *Store) GetTopPeersID() (*[]common.Address, error) {
	var peersID wire.PeersID
	if err := s.get(s.top10PeersID, []byte{0}, &peersID); err != nil {
		return nil, err
	}

	addresses := make([]common.Address, len(peersID.ID))

	// TODO: too slow solution
	for _, id := range peersID.ID {
		addresses = append(addresses, common.HexToAddress(id))
	}

	return &addresses, nil
}

// SetPeer stores peer.
func (s *Store) SetPeer(peer *Peer) error {
	w := peer.ToWire()
	return s.set(s.peers, peer.ID.Bytes(), w)
}

// GetPeerInfo returns stored peer info.
// Result is a ready gRPC message.
func (s *Store) GetPeerInfo(id common.Address) (*wire.PeerInfo, error) {
	var peer wire.PeerInfo
	if err := s.get(s.peers, id.Bytes(), &peer); err != nil {
		return nil, err
	}

	return &peer, nil
}

// GetPeer returns stored peer.
func (s *Store) GetPeer(id common.Address) (*Peer, error) {
	w, err := s.GetPeerInfo(id)
	if err != nil {
		return nil, err
	}

	return WireToPeer(w), nil
}

/*
 * Utils:
 */

func (s *Store) set(table kvdb.Database, key []byte, val proto.Message) error {
	var pbf proto.Buffer

	if err := pbf.Marshal(val); err != nil {
		panic(err)
	}

	if err := table.Put(key, pbf.Bytes()); err != nil {
		panic(err)
	}

	return nil
}

func (s *Store) get(table kvdb.Database, key []byte, to proto.Message) error {
	buf, err := table.Get(key)
	if err != nil {
		panic(err)
	}
	if buf == nil {
		return nil
	}

	if err = proto.Unmarshal(buf, to); err != nil {
		panic(err)
	}

	return nil
}
