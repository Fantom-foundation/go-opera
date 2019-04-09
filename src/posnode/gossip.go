package posnode

import (
	"context"
	"fmt"
	"sync"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/posnode/api"
)

// gossip is a pool of gossiping processes.
type gossip struct {
	tickets chan struct{}

	sync.Mutex
}

func (g *gossip) addTicket() {
	g.Lock()
	defer g.Unlock()
	if g.tickets != nil {
		g.tickets <- struct{}{}
	}
}

// StartGossip starts gossiping.
func (n *Node) StartGossip(threads int) {
	if n.gossip.tickets != nil {
		return
	}

	n.initPeers()

	n.gossip.tickets = make(chan struct{}, threads)
	for i := 0; i < threads; i++ {
		n.gossip.addTicket()
	}

	go n.gossiping()

	n.log.Info("gossip started")
}

// StopGossip stops gossiping.
func (n *Node) StopGossip() {
	n.gossip.Lock()
	defer n.gossip.Unlock()
	close(n.gossip.tickets)
	n.gossip.tickets = nil

	n.log.Info("gossip stopped")
}

// gossiping is a infinity gossip process.
func (n *Node) gossiping() {
	for range n.gossip.tickets {
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
		select {} // for cpu rest
		return
	}
	defer n.FreePeer(peer)

	client, err := n.ConnectTo(peer)
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), clientTimeout)
	defer cancel()

	knownHeights := n.knownEvents()
	unknownHeights, err := client.SyncEvents(ctx, knownHeights)
	if err != nil {
		n.ConnectFail(peer, err)
		return
	}
	n.ConnectOK(peer)

	peers2discovery := make(map[hash.Peer]struct{})

	toDelete, toDownload := n.addToDownloads(&unknownHeights.Lasts)
	defer n.deleteFromDownloads(toDelete)

	for hex, height := range *toDownload {
		req := &api.EventRequest{
			PeerID: hex,
			Index:  height,
		}

		creator := hash.HexToPeer(hex)

		event := n.getEvent(client, peer, req)
		if event == nil {
			return
		}

		if event.Creator != creator || event.Index != height {
			n.ConnectFail(peer, fmt.Errorf("bad GetEvent() response"))
			return
		}

		n.SaveNewEvent(event)

		n.checkParents(client, peer, &event.Parents)

		peers2discovery[creator] = struct{}{}
	}
	n.ConnectOK(peer)

	// check peers from events
	for p := range peers2discovery {
		n.CheckPeerIsKnown(peer.ID, peer.Host, p)
	}
}

func (n *Node) checkParents(client api.NodeClient, peer *Peer, parents *hash.Events) {
	toDownload := n.addParentToDownloads(parents)
	defer n.deleteParentFromDownloads(toDownload)

	for p := range *toDownload {
		var req api.EventRequest
		req.Hash = p.Bytes()

		event := n.getEvent(client, peer, &req)
		if event == nil {
			return
		}

		// Add event to store
		n.store.SetEvent(event)
		n.store.SetEventHash(event.Creator, event.Index, event.Hash())

		// We don't need to store heights here
	}
}

// Get event & check sign
func (n *Node) getEvent(client api.NodeClient, peer *Peer, requets *api.EventRequest) *inter.Event {
	ctx, cancel := context.WithTimeout(context.Background(), clientTimeout)
	w, err := client.GetEvent(ctx, requets)
	cancel()

	if err != nil {
		n.ConnectFail(peer, err)
		return nil
	}

	event := inter.WireToEvent(w)

	// Check event sign
	creator := n.store.GetPeer(event.Creator)
	if creator == nil {
		return nil
	}

	if !event.Verify(creator.PubKey) {
		n.ConnectFail(peer, fmt.Errorf("falsity GetEvent() response"))
		return nil
	}

	return event
}

// SaveNewEvent writes event to store and indexes.
// Note: event should be last from its creator.
func (n *Node) SaveNewEvent(e *inter.Event) {
	n.store.SetEvent(e)
	n.store.SetEventHash(e.Creator, e.Index, e.Hash())
	n.store.SetPeerHeight(e.Creator, e.Index)
}

// knownEventsReq makes request struct with event heights of top peers.
func (n *Node) knownEvents() *api.KnownEvents {
	peers := n.peers.Snapshot()
	peers = append(peers, n.ID)

	res := &api.KnownEvents{
		Lasts: make(map[string]uint64, len(peers)),
	}

	for _, id := range peers {
		h := n.store.GetPeerHeight(id)
		res.Lasts[id.Hex()] = h
	}

	return res
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
	a := n.peers.attrOf(n.peers.top[i])
	b := n.peers.attrOf(n.peers.top[j])

	if a.LastSuccess.After(a.LastFail) && !b.LastSuccess.After(b.LastFail) {
		return true
	}

	if a.LastFail.After(a.LastSuccess) && b.LastFail.After(b.LastSuccess) {
		if a.LastFail.Before(b.LastFail) {
			return true
		}
	}

	if a.LastSuccess.After(b.LastSuccess) {
		return true
	}

	return false
}
