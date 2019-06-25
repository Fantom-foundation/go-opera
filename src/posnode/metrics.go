package posnode

import (
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

	// count of downloaded events
	countDownloadedEvents = metrics.RegisterCounter("count_downloaded_events", nil)

	// count total events
	countTotalEvents = metrics.RegisterCounter("count_total_events", nil)
)
