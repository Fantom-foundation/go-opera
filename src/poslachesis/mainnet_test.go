package lachesis

import (
	"fmt"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
)

// LachesisNetworkRing starts lachesis network with initial ring topology.
func LachesisNetworkRing(count int, balance uint64) []*Lachesis {
	if count < 1 {
		return nil
	}

	res, _ := makeNetwork(count, balance)

	// init peers ring
	for i := 0; i < count; i++ {
		node := res[i].node

		j := (i + 1) % count
		peer := res[j].node

		node.CheckPeerIsKnown(hash.EmptyPeer, peer.Host(), nil)
	}

	return res
}

// LachesisNetworkStar starts lachesis network with initial star topology.
func LachesisNetworkStar(count int, balance uint64) []*Lachesis {
	if count < 1 {
		return nil
	}

	res, _ := makeNetwork(count, balance)

	// init peers star
	for i := 1; i < count; i++ {
		node := res[i].node

		peer := res[0].node

		node.CheckPeerIsKnown(hash.EmptyPeer, peer.Host(), nil)
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

	// start all
	for _, l := range ll {
		l.Start(genesis)
	}

	return ll, genesis
}
