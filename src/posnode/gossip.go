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
}

// StopGossip stops gossiping.
func (n *Node) StopGossip() {
	n.gossip.Lock()
	defer n.gossip.Unlock()
	close(n.gossip.tickets)
	n.gossip.tickets = nil
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

	for hex, height := range unknownHeights.Lasts {
		req := &api.EventRequest{
			PeerID: hex,
		}
		creator := hash.HexToPeer(hex)
		last := n.store.GetPeerHeight(creator)
		for i := last + 1; i <= height; i++ {
			key := hex + string(i)

			// Check event in downloads before call GetEvent
			if alreadyExist := n.checkDownloads(key); alreadyExist {
				continue
			}

			// Add record about event to downloads before get it
			n.addToDownloads(key)

			req.Index = i

			ctx, cancel := context.WithTimeout(context.Background(), clientTimeout)
			w, err := client.GetEvent(ctx, req)
			cancel()

			// Delete record about event from downloads even if we get error
			n.deleteFromDownloads(key)

			if err != nil {
				n.ConnectFail(peer, err)
				return
			}

			event := inter.WireToEvent(w)

			if event.Creator != creator || event.Index != i {
				n.ConnectFail(peer, fmt.Errorf("bad GetEvent() response"))
				return
			}

			// check event sign
			signer := n.store.GetPeer(creator)
			if signer == nil {
				return
			}
			if !event.Verify(peer.PubKey) {
				n.ConnectFail(peer, fmt.Errorf("falsity GetEvent() response"))
				return
			}

			n.SaveNewEvent(event)

			n.checkParents(client, peer, event.Parents)

			peers2discovery[creator] = struct{}{}
		}
	}
	n.ConnectOK(peer)

	// check peers from events
	for p := range peers2discovery {
		n.CheckPeerIsKnown(peer.ID, peer.Host, p)
	}
}

func (n *Node) checkParents(client api.NodeClient, peer *Peer, parents hash.Events) {
	for p := range parents {

		// Check parent in store
		if n.store.GetEvent(p) != nil {
			continue
		}

		// Check event in downloads before call GetEvent
		if alreadyExist := n.checkParentDownloads(p); alreadyExist {
			continue
		}

		// Add record about event to downloads before get it
		n.addParentToDownloads(p)

		var req api.EventRequest
		req.Hash = p.Bytes()

		ctx, cancel := context.WithTimeout(context.Background(), clientTimeout)
		w, err := client.GetEvent(ctx, &req)
		cancel()

		// Delete record about event from downloads even if we get error
		n.deleteParentFromDownloads(p)

		if err != nil {
			n.ConnectFail(peer, err)
			return
		}

		event := inter.WireToEvent(w)

		// Check event sign
		peer := n.store.GetPeer(event.Creator)
		if peer == nil {
			return
		}

		if !event.Verify(peer.PubKey) {
			n.ConnectFail(peer, fmt.Errorf("falsity GetEvent() response"))
			return
		}

		// Add event to store
		n.store.SetEvent(event)
		n.store.SetEventHash(event.Creator, event.Index, event.Hash())

		// We don't need to store heights here
	}
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
// TODO: test it
func (n *gossipEvaluation) Less(i, j int) bool {
	a := n.peers.attrOf(n.peers.top[i])
	b := n.peers.attrOf(n.peers.top[j])

	if a.LastSuccess.After(a.LastFail) && !b.LastSuccess.After(b.LastFail) {
		return true
	}

	if a.LastSuccess.Before(b.LastSuccess) {
		return true
	}

	if a.LastFail.After(b.LastFail) {
		return true
	}

	return false
}
