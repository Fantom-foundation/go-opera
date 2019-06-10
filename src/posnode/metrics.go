package posnode

import (
	"github.com/Fantom-foundation/go-lachesis/src/metrics"
)

var (
	countNodePeersTop = metrics.NewRegisteredCounter("count_node_peers_top", nil)
)
