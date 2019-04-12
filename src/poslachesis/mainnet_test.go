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

	res, genesis := makeNetwork(count, balance)

	// init peers ring
	for i := 0; i < count; i++ {
		node := res[i].nodeStore

		j := (i + 1) % count
		peer := res[j].node

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

// LachesisNetworkStar starts lachesis network with initial star topology.
func LachesisNetworkStar(count int, balance uint64) []*Lachesis {
	if count < 1 {
		return nil
	}

	res, genesis := makeNetwork(count, balance)

	// init peers star
	for i := 1; i < count; i++ {
		node := res[i].nodeStore

		peer := res[0].node

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

func makeNetwork(count int, balance uint64) ([]*Lachesis, map[hash.Peer]uint64) {
	ll := make([]*Lachesis, count)
	genesis := make(map[hash.Peer]uint64, count)

	// create all
	for i := 0; i < count; i++ {
		host := fmt.Sprintf("node_%d", i)
		lachesis := NewForTests(nil, host)
		genesis[lachesis.node.ID] = balance

		ll[i] = lachesis
	}

	return ll, genesis
}
