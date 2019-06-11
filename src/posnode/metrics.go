package posnode

import (
	"github.com/Fantom-foundation/go-lachesis/src/metrics"
)

var (
	countNodePeersTop = metrics.RegisterCounter("count_node_peers_top", nil)
)
