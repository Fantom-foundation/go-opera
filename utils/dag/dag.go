package dag

import (
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"gonum.org/v1/gonum/graph/encoding/dot"

	"github.com/Fantom-foundation/go-opera/gossip"
)

func Graph(db *gossip.Store, from, to idx.Epoch) dot.Graph {
	return &dagReader{
		db:        db,
		epochFrom: from,
		epochTo:   to,
	}
}
