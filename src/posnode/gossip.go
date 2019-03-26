package posnode

import (
	"context"
	"sync"
	"time"

	"google.golang.org/grpc"

	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/posnode/wire"
)

// TODO: 
// Change to return err or panic later 

// if err != nil {
// 	n.log().Error(err)
// }

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
	n.log().Debug("gossip +")

	// Get top peers id
	ids, err := n.store.GetTopPeersID()
	if err != nil {
		n.log().Error(err)
	}

	if len(*ids) == 0 {
		// TODO: add bootstrap func for Top10PeersID & them addresses
	}

	// Check already connected nodes
	var selectedPeer common.Address
	for _, id := range *ids {
		if !n.connectedPeers[id] {
			selectedPeer = id
		}
	}

	// Get peer
	peer, err := n.store.GetPeer(selectedPeer)
	if err != nil {
		n.log().Error(err)
	}

	// Connect
	var conn *grpc.ClientConn
	conn, err = grpc.DialContext(context.Background(), peer.NetAddr, grpc.WithInsecure())
	if err != nil {
		n.log().Error(err)
	}
	defer conn.Close()

	client := wire.NewNodeClient(conn)

	// Mark peer as connected
	n.connectedPeers[selectedPeer] = true

	// Get events from peer
	peers := n.syncWithPeer(client, peer)

	<-time.After(time.Second / 2)

	// Check peers from events
	for p := range peers {
		n.CheckPeerIsKnown(n.ID, p)
	}

	// Mark connection as close
	n.connectedPeers[selectedPeer] = false

	n.log().Debug("gossip -")
}

func (n *Node) syncWithPeer(client wire.NodeClient, peer *Peer) map[common.Address]bool {

	// Send known heights
	unknownHeights, err := n.sendSyncEventsRequest(client, n.knownHeights)
	if err != nil {
		n.log().Error(err)
	}

	// Collect peers from each event
	var peers map[common.Address]bool

	// Get unknown events by heights
	for pID, height := range unknownHeights.Lasts {
		if n.knownHeights[pID] < height {
			for i := n.knownHeights[pID] + 1; i <= height; i++ {
				
				var req wire.EventRequest
				req.PeerID = pID
				req.Index = i

				event, err := n.sendGetEventRequest(client, &req)
				if err != nil {
					n.log().Error(err)
				}

				address := common.BytesToAddress(event.Creator)
				peers[address] = false
			}

			n.knownHeights[pID] = height
		}
	}

	return peers
}

func (n *Node) sendSyncEventsRequest(client wire.NodeClient, knownHeights map[string]uint64) (*wire.KnownEvents, error) {
	unknown, err := client.SyncEvents(context.Background(), &wire.KnownEvents{Lasts: knownHeights})
	if err != nil {
		n.log().Error(err)
	}

	return unknown, nil
}

func (n *Node) sendGetEventRequest(client wire.NodeClient, req *wire.EventRequest) (*wire.Event, error) {
	event, err := client.GetEvent(context.Background(), req)
	if err != nil {
		n.log().Error(err)
	}

	return event, nil
}
