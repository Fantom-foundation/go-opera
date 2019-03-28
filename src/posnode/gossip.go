package posnode

import (
	"context"
	"sync"

	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/posnode/wire"
)

// gossip is a pool of gossiping processes.
type gossip struct {
	Tickets chan struct{}
	Sync    sync.Mutex
}

func (g *gossip) addTicket() {
	g.Sync.Lock()
	defer g.Sync.Unlock()
	if g.Tickets != nil {
		g.Tickets <- struct{}{}
	}
}

// StartGossip starts gossiping.
// It should be called once.
func (n *Node) StartGossip(threads int) {
	n.gossip.Tickets = make(chan struct{}, threads)
	for i := 0; i < threads; i++ {
		n.gossip.addTicket()
	}

	go n.gossiping()
}

// StopGossip stops gossiping.
// It should be called once.
func (n *Node) StopGossip() {
	n.gossip.Sync.Lock()
	defer n.gossip.Sync.Unlock()
	close(n.gossip.Tickets)
	n.gossip.Tickets = nil
}

// gossiping is a infinity gossip process.
func (n *Node) gossiping() {
	for range n.gossip.Tickets {
		go func() {
			defer n.gossip.addTicket()
			n.gossipOnce()
		}()
	}
}

func (n *Node) gossipOnce() {
	ids := n.store.GetTopPeers()

	// Check already connected nodes
	var selectedPeer common.Address
	for _, id := range ids {
		// Check for unconnected peer & not self connected
		if !n.connectedPeers.Load(id) && n.ID != id {
			selectedPeer = id
		}
	}

	// If don't have free peer -> return without error
	if (selectedPeer == common.Address{0}) {
		return
	}

	// Get peer
	peer := n.store.GetPeer(selectedPeer)
	if peer == nil {
		return // If we have peer's ID but does not have a peer's data -> just return
	}

	// Connect
	ctx, cancel := context.WithTimeout(context.Background(), clientTimeout)
	defer cancel()

	client, err := n.ConnectTo(ctx, peer.NetAddr)
	if err != nil {
		n.log.Error(err)
		return // if refused -> return without error
	}

	n.log.Debug("connect to ", peer.NetAddr)

	// Mark peer as connected
	n.connectedPeers.Store(selectedPeer, true)

	// Get events from peer
	peers := n.syncWithPeer(client, peer)

	// Check peers from events
	for p := range peers {
		n.CheckPeerIsKnown(peer.ID, p, peer.NetAddr)
	}

	// Mark connection as close
	n.connectedPeers.Store(selectedPeer, false)
}

func (n *Node) syncWithPeer(client wire.NodeClient, peer *Peer) map[common.Address]bool {
	knownHeights := n.store_GetHeights()

	// Send known heights -> get unknown
	unknownHeights, err := client.SyncEvents(context.Background(), &wire.KnownEvents{Lasts: (*knownHeights).Lasts})
	if err != nil {
		n.log.Error(err)
		return map[common.Address]bool{} // if connection refused -> return empty map without error
	}

	// Collect peers from each event
	var peers map[common.Address]bool

	// Get unknown events by heights
	for pID, height := range unknownHeights.Lasts {
		if (*knownHeights).Lasts[pID] < height {
			for i := (*knownHeights).Lasts[pID] + 1; i <= height; i++ {

				var req wire.EventRequest
				req.PeerID = pID
				req.Index = i

				event, err := client.GetEvent(context.Background(), &req)
				if err != nil {
					n.log.Error(err)
					return map[common.Address]bool{} // if connection refused -> return empty map without error
				}

				address := common.BytesToAddress(event.Creator)
				peers[address] = false
			}

			(*knownHeights).Lasts[pID] = height
		}
	}

	n.store_SetHeights(knownHeights)

	return peers
}

// NOTE: temporary decision
func (n *Node) store_GetHeights() *wire.KnownEvents {
	res := &wire.KnownEvents{
		Lasts: make(map[string]uint64),
	}

	for _, id := range n.store.GetKnownPeers() {
		h := n.store.GetPeerHeight(id)
		res.Lasts[id.Hex()] = h
	}

	return res
}

// NOTE: temporary decision
func (n *Node) store_SetHeights(w *wire.KnownEvents) {
	ids := make([]common.Address, 0)

	for str, h := range w.Lasts {
		id := common.HexToAddress(str)
		ids = append(ids, id)
		n.store.SetPeerHeight(id, h)
	}
	n.store.SetKnownPeers(ids)
}
