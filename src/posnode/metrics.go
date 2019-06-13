package posnode

import (
	"time"

	"github.com/Fantom-foundation/go-lachesis/src/metrics"
)

var (
	// count of peers in top
	countNodePeersTop = metrics.RegisterCounter("count_node_peers_top", nil)

	// count of emmited events by node
	countEmittedEvents = metrics.RegisterCounter("count_emitted_events", nil)

	// count of active connections
	countConnections = metrics.RegisterCounter("count_connections", nil)

	// last successful gossip time (unix).
	lastSuccessGossipTime = metrics.RegisterGauge("last_success_gossip_time", nil)

	// last fail gossip time (unix).
	lastFailGossipTime = metrics.RegisterGauge("last_fail_gossip_time", nil)

	// last successful discovery time (unix).
	lastSuccessDiscoveryTime = metrics.RegisterGauge("last_success_discovery_time", nil)

	// last fail discovery time (unix).
	lastFailDiscoveryTime = metrics.RegisterGauge("last_fail_discovery_time", nil)
)

func (n *Node) gossipSuccess(p *Peer) {
	lastSuccessGossipTime.Update(time.Now().Unix())

	n.ConnectOK(p)
}

func (n *Node) gossipFail(p *Peer, err error) {
	lastFailGossipTime.Update(time.Now().Unix())

	n.ConnectFail(p, err)
}

func (n *Node) discoverySuccess(p *Peer) {
	lastSuccessDiscoveryTime.Update(time.Now().Unix())

	n.ConnectOK(p)
}

func (n *Node) discoveryFail(p *Peer, err error) {
	lastFailDiscoveryTime.Update(time.Now().Unix())

	n.ConnectFail(p, err)
}
