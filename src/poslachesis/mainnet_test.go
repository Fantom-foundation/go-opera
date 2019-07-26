package lachesis

import (
	"fmt"
	"time"
)

// LachesisNetworkRing starts lachesis network with initial ring topology.
func LachesisNetworkRing(count int) []*Lachesis {
	if count < 1 {
		return nil
	}

	res := makeNetwork("ring", count)

	// init peers ring
	for i := 0; i < count; i++ {
		node := res[i].node

		j := (i + 1) % count
		peer := res[j].node

		node.CheckPeerIsKnown(peer.Host(), nil)
	}

	return res
}

// LachesisNetworkStar starts lachesis network with initial star topology.
func LachesisNetworkStar(count int) []*Lachesis {
	if count < 1 {
		return nil
	}

	res := makeNetwork("star", count)

	// init peers star
	for i := 1; i < count; i++ {
		node := res[i].node

		peer := res[0].node

		node.CheckPeerIsKnown(peer.Host(), nil)
	}

	return res
}

func makeNetwork(pref string, count int) []*Lachesis {
	net, keys := FakeNet(count)

	conf := DefaultConfig()
	conf.Net = net
	conf.Node.MinEmitInterval = 1 * time.Second
	conf.Node.MaxEmitInterval = 4 * time.Second

	ll := make([]*Lachesis, count)

	// create all
	for i, key := range keys {
		host := fmt.Sprintf("%s%s", pref, string('A'+i))
		ll[i] = NewForTests(nil, host, key, conf)
	}

	// start all
	for _, l := range ll {
		l.Start()
	}

	return ll
}
