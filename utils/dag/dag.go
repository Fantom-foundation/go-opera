package dag

import (
	"github.com/Fantom-foundation/lachesis-base/inter/idx"

	"github.com/Fantom-foundation/go-opera/gossip"
	"github.com/Fantom-foundation/go-opera/integration"
	"github.com/Fantom-foundation/go-opera/utils/dag/dot"
)

func Graph(db *gossip.Store, cfg integration.Configs, from, to idx.Epoch) *dot.Graph {
	g := readDagGraph(db, cfg, from, to)

	return g
}
