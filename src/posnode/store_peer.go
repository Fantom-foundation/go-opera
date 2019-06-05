package posnode

import (
	"github.com/golang/protobuf/proto"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/posnode/api"
)

// BootstrapPeers stores peer list.
func (s *Store) BootstrapPeers(peers ...*Peer) {
	if len(peers) < 1 {
		return
	}

	// save peers
	batch := s.Peers.NewBatch()
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

// SetPeer stores peer.
func (s *Store) SetPeer(peer *Peer) {
	info := peer.ToWire()
	s.SetWirePeer(peer.ID, info)
}

// GetPeer returns stored peer.
func (s *Store) GetPeer(id hash.Peer) *Peer {
	w := s.GetWirePeer(id)
	return WireToPeer(w)
}

// SetWirePeer stores peer info.
func (s *Store) SetWirePeer(id hash.Peer, info *api.PeerInfo) {
	s.set(s.Peers, id.Bytes(), info)
}

// GetWirePeer returns stored peer info.
// Result is a ready gRPC message.
func (s *Store) GetWirePeer(id hash.Peer) *api.PeerInfo {
	w, _ := s.get(s.Peers, id.Bytes(), &api.PeerInfo{}).(*api.PeerInfo)
	return w
}

// SetTopPeers stores peers.top.
func (s *Store) SetTopPeers(ids []hash.Peer) {
	var key = []byte("current")
	w := IDsToWire(ids)
	s.set(s.PeersTop, key, w)
}

// GetTopPeers returns peers.top.
func (s *Store) GetTopPeers() []hash.Peer {
	var key = []byte("current")
	w, _ := s.get(s.PeersTop, key, &api.PeerIDs{}).(*api.PeerIDs)
	return WireToIDs(w)
}

// SetPeerHeight stores last event index of peer.
func (s *Store) SetPeerHeight(id hash.Peer, height uint64) {
	if err := s.PeerHeights.Put(id.Bytes(), intToBytes(height)); err != nil {
		s.Fatal(err)
	}
}

// GetPeerHeight returns last event index of peer.
func (s *Store) GetPeerHeight(id hash.Peer) uint64 {
	buf, err := s.PeerHeights.Get(id.Bytes())
	if err != nil {
		s.Fatal(err)
	}
	if buf == nil {
		return 0
	}

	return bytesToInt(buf)
}
