package posnode

import (
	"fmt"
	"sync"
	"time"
)

// gossip is a pool of gossiping processes.
type gossip struct {
	Tickets chan struct{}
	Sync    sync.Mutex
}

func (g *gossip) addTicket() {
	g.Sync.Lock()
	if g.Tickets != nil {
		g.Tickets <- struct{}{}
	}
	g.Sync.Unlock()
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
	close(n.gossip.Tickets)
	n.gossip.Tickets = nil
	n.gossip.Sync.Unlock()
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
	// TODO: implement it (select peer, connect, sync with peer, get new events)
	fmt.Println("gossip +")
	<-time.After(time.Second / 2)
	fmt.Println("gossip -")
}
