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
			n.syncWithPeer()
		}()
	}
}

func (n *Node) syncWithPeer() {
	peer := n.NextForGossip()
	if peer == nil {
		n.log.Warn("no candidate for gossip")
		// TODO: wait for timeout here
		return
	}
	defer n.FreePeer(peer)

	// Connect
	ctx, cancel := context.WithTimeout(context.Background(), clientTimeout)
	defer cancel()

	client, err := n.ConnectTo(ctx, peer.Host)
	if err != nil {
		n.log.Warn(err)
		return
	}

	knownHeights := n.store_GetHeights()

	// Send known heights -> get unknown
	unknownHeights, err := client.SyncEvents(context.Background(), &wire.KnownEvents{Lasts: knownHeights.Lasts})
	if err != nil {
		n.log.Warn(err)
		return
	}

	// Collect peers from each event
	var peers map[common.Address]bool

	// Get unknown events by heights
	for pID, height := range unknownHeights.Lasts {
		if knownHeights.Lasts[pID] < height {
			for i := knownHeights.Lasts[pID] + 1; i <= height; i++ {

				var req wire.EventRequest
				req.PeerID = pID
				req.Index = i

				event, err := client.GetEvent(context.Background(), &req)
				if err != nil {
					n.log.Warn(err)
					return
				}

				address := common.BytesToAddress(event.Creator)
				peers[address] = false
			}

			knownHeights.Lasts[pID] = height
		}
	}

	n.store_SetHeights(knownHeights)

	// Check peers from events
	for p := range peers {
		n.CheckPeerIsKnown(peer.ID, p, peer.Host)
	}
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

/*
 * evaluation function for gossip
 */

// gossipEvaluation implements sort.Interface.
type gossipEvaluation Node

// Len is the number of elements in the collection.
func (n *gossipEvaluation) Len() int {
	return len(n.peers.top)
}

// Swap swaps the elements with indexes i and j.
func (n *gossipEvaluation) Swap(i, j int) {
	n.peers.top[i], n.peers.top[j] = n.peers.top[j], n.peers.top[i]
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (n *gossipEvaluation) Less(i, j int) bool {
	a := n.store.GetPeer(n.peers.top[i])
	b := n.store.GetPeer(n.peers.top[j])
	if a == nil || b == nil {
		panic("unsaved peer detected in node peers")
	}

	// TODO: implement a vs b comparing
	return false
}
