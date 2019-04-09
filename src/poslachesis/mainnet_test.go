package lachesis

import (
	"fmt"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/posnode"
)

// LachesisNetworkRing starts lachesis network with initial ring topology.
func LachesisNetworkRing(count int, balance uint64) []*Lachesis {
	if count < 1 {
		return nil
	}

	res := make([]*Lachesis, count)
	genesis := make(map[hash.Peer]uint64, count)

	// create all
	for i := 0; i < count; i++ {
		host := fmt.Sprintf("node_%d", i)
		lachesis := NewForTests(nil, host)
		genesis[lachesis.Node.ID] = balance

		res[i] = lachesis
	}

	// init peers
	for i := 0; i < count; i++ {
		node := res[i].nodeStore

		j := (i + 1) % count
		peer := res[j].Node

		node.BootstrapPeers(&posnode.Peer{
			ID:     peer.ID,
			PubKey: peer.PubKey(),
			Host:   peer.Host(),
		})
	}

	// start all
	for _, l := range res {
		l.Start(genesis)
	}

	return res
}
