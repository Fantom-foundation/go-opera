package posnode

import (
	"fmt"
	"sync"
	"time"

	"github.com/Fantom-foundation/go-lachesis/src/common"
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
	for _ = range n.gossip.Tickets {
		go func() {
			defer n.gossip.addTicket()
			n.gossipOnce()
		}()
	}
}

func (n *Node) gossipOnce() {
	// TODO: implement it (select peer, connect, sync with peer, get new events, n.CheckPeerIsKnown())
	fmt.Println("gossip +")
	<-time.After(time.Second / 2)

	n.CheckPeerIsKnown(common.Address{}, common.Address{})

	fmt.Println("gossip -")
}
