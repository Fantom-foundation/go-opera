package posnode

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/posnode/api"
)

const (
	gossipIdle = time.Second * 5
)

// gossip is a pool of gossiping processes.
type gossip struct {
	tickets chan struct{}

	sync.Mutex
}

func (g *gossip) freeTicket() {
	g.Lock()
	defer g.Unlock()

	if g.tickets != nil {
		g.tickets <- struct{}{}
	}
}

// StartGossip starts gossiping.
func (n *Node) StartGossip(threads int) {
	n.gossip.Lock()
	defer n.gossip.Unlock()

	if n.gossip.tickets != nil {
		return
	}

	n.initPeers()

	n.gossip.tickets = make(chan struct{}, threads)
	for i := 0; i < threads; i++ {
		n.gossip.tickets <- struct{}{}
	}

	go n.gossiping(n.gossip.tickets)

	n.Info("gossip started")
}

// StopGossip stops gossiping.
func (n *Node) StopGossip() {
	n.gossip.Lock()
	defer n.gossip.Unlock()

	if n.gossip.tickets == nil {
		return
	}

	close(n.gossip.tickets)
	n.gossip.tickets = nil

	n.Info("gossip stopped")
}

// gossiping is a infinity gossip process.
func (n *Node) gossiping(tickets chan struct{}) {
	for range tickets {
		go func() {
			defer n.gossip.freeTicket()
			peer := n.NextForGossip()
			if peer != nil {
				defer n.FreePeer(peer)
				n.syncWithPeer(peer)
			} else {
				n.Warn("no candidate for gossip")
			}
			time.Sleep(gossipIdle)
		}()
	}

}

func (n *Node) syncWithPeer(peer *Peer) {
	peers2discovery := make(map[hash.Peer]struct{})

	discovery := func(peers2discovery map[hash.Peer]struct{}) {
		// check peers from events
		for p := range peers2discovery {
			n.Infof("peers2discovery %s", p.String())
			n.CheckPeerIsKnown(peer.Host, &p)
		}

		// Clean outdated data about peers.
		n.trimHosts(n.conf.TopPeersCount*4, n.conf.TopPeersCount*3)
	}

	defer discovery(peers2discovery)

	client, free, fail, err := n.ConnectTo(peer)
	if err != nil {
		n.Error(err)
		return
	}
	defer free()

	n.Debugf("gossip with peer %s", peer.ID.String())

	sf := n.superFrame()
	unknowns, err := n.compareKnownEvents(client, peer, sf)
	if err != nil {
		fail(err)
		return
	}
	if unknowns == nil {
		return
	}

	parents := hash.Events{}

	toDownload := n.lockFreeHeights(sf, unknowns)
	defer n.unlockFreeHeights(sf, toDownload)

	for creator, interval := range toDownload {
		req := &api.EventRequest{
			PeerID: creator.Hex(),
		}

		peers2discovery[creator] = struct{}{}

		for i := interval.from; i <= interval.to; i++ {
			req.Seq = uint64(i)

			event, err := n.downloadEvent(client, peer, req)
			if err != nil {
				fail(err)
				return
			}
			if event == nil {
				return
			}

			parents.Add(event.Parents.Slice()...)
		}
	}
	n.gossipSuccess(peer)

	n.checkParents(client, peer, parents)
}

func (n *Node) checkParents(client api.NodeClient, peer *Peer, parents hash.Events) {
	toDownload := n.lockNotDownloaded(parents)
	defer n.unlockDownloaded(toDownload)

	n.Info("check parents")

	for e := range toDownload {
		if e == hash.ZeroEvent {
			continue
		}

		var req api.EventRequest
		req.Hash = e.Bytes()

		event, err := n.downloadEvent(client, peer, &req)
		if err != nil {
			n.Warnf("download parent event error: %s", err.Error())
		}

		if event == nil {
			n.Warn("download parent event error: Event is nil")
		}
	}
}

func (n *Node) compareKnownEvents(client api.NodeClient, peer *Peer, sf idx.SuperFrame) (heights, error) {
	knowns := n.knownEvents(sf)

	n.Debugf("%s knows %v at SF%d", n.ID.String(), knowns, sf)

	req := &api.KnownEvents{
		SuperFrameN: uint64(sf),
		Lasts:       knowns.ToWire(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), n.conf.ClientTimeout)
	defer cancel()

	id, ctx := api.ServerPeerID(ctx)

	resp, err := client.SyncEvents(ctx, req)
	if err != nil {
		n.gossipFail(peer, err)
		return nil, err
	}

	if *id != peer.ID {
		// TODO: skip or continue gossiping with peer id ?
	}

	n.Debugf("%s knows %v", peer.ID.String(), resp.Lasts)

	if resp.SuperFrameN != req.SuperFrameN {
		err = fmt.Errorf("unexpected super-frame-index %d, expected %d", resp.SuperFrameN, req.SuperFrameN)
		n.gossipFail(peer, err)
		return nil, err
	}

	if req.SuperFrameN != 0 {
		if len(resp.Lasts) != len(req.Lasts) {
			err = fmt.Errorf("unexpected super-frame-members are different")
			return nil, err
		}
		for member := range req.Lasts {
			if _, ok := resp.Lasts[member]; !ok {
				err = fmt.Errorf("unexpected super-frame-members are different")
				return nil, err
			}
		}
	}

	lasts := WireToHeights(resp.Lasts)

	res := lasts.Exclude(knowns)

	n.Debugf("DIFF %s between %s and %s", res.String(), peer.ID.String(), n.ID.String())

	n.gossipSuccess(peer)
	return res, nil
}

// downloadEvent downloads event.
func (n *Node) downloadEvent(client api.NodeClient, peer *Peer, req *api.EventRequest) (*inter.Event, error) {
	ctx, cancel := context.WithTimeout(context.Background(), n.conf.ClientTimeout)
	defer cancel()

	id, ctx := api.ServerPeerID(ctx)

	w, err := client.GetEvent(ctx, req)
	if err != nil {
		n.gossipFail(peer, err)
		return nil, err
	}

	if *id != peer.ID {
		// TODO: skip or continue gossiping with peer id ?
	}

	if req.Hash == nil {
		if w.Creator != req.PeerID || w.Seq != req.Seq {
			n.gossipFail(peer, fmt.Errorf("bad GetEvent() response"))
			return nil, nil
		}
	}

	event := inter.WireToEvent(w)

	// check that event creator is super frame member
	isKnownMember := false
	members := n.consensus.SuperFrameMembers(n.superFrame())
	for _, m := range members {
		if event.Creator == m {
			isKnownMember = true
			break
		}
	}
	if !isKnownMember {
		err = fmt.Errorf("falsity GetEvent() response")
		return nil, err
	}

	// check event sign
	creator := n.store.GetPeer(event.Creator)
	if creator == nil {
		return nil, nil
	}
	if !event.Verify(creator.PubKey) {
		err = fmt.Errorf("falsity GetEvent() response")
		n.gossipFail(peer, err)
		return nil, err
	}
	n.Infof("downloaded event %s", event.Hash().String())

	n.onNewEvent(event)

	countDownloadedEvents.Inc(1)

	return event, nil
}

// knownEventsReq returns event heights of requested super-frame.
func (n *Node) knownEvents(sf idx.SuperFrame) heights {
	if sf > n.superFrame() {
		return heights{}
	}

	var peers []hash.Peer

	if n.consensus == nil {
		peers = n.peers.Snapshot()
	} else {
		peers = n.consensus.SuperFrameMembers(sf)
	}

	return n.peersWithHeight(sf, peers)
}

func (n *Node) peersWithHeight(sf idx.SuperFrame, peers []hash.Peer) heights {
	peers = append(peers, n.ID)

	res := make(heights, len(peers))
	for _, id := range peers {
		h := n.store.GetPeerHeight(id, sf)
		res[id] = interval{to: h}
	}

	return res
}

func (n *Node) gossipSuccess(p *Peer) {
	lastSuccessGossipTime.Update(time.Now().Unix())

	n.ConnectOK(p)
}

func (n *Node) gossipFail(p *Peer, err error) {
	lastFailGossipTime.Update(time.Now().Unix())

	n.ConnectFail(p, err)
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
	a := n.peers.attrByID(n.peers.top[i]).Host
	b := n.peers.attrByID(n.peers.top[j]).Host

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
