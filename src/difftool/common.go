package difftool

import (
	"crypto/ecdsa"
	"fmt"
	"math/rand"

	"github.com/sirupsen/logrus"

	"github.com/andrecronje/lachesis/src/crypto"
	"github.com/andrecronje/lachesis/src/dummy"
	"github.com/andrecronje/lachesis/src/net"
	"github.com/andrecronje/lachesis/src/node"
	"github.com/andrecronje/lachesis/src/peers"
	"github.com/andrecronje/lachesis/src/poset"
)

type NodeList map[*ecdsa.PrivateKey]*node.Node

func NewNodeList(count int, logger *logrus.Logger) NodeList {
	nodes := make(NodeList, count)
	participants := peers.NewPeers()

	for i := 0; i < count; i++ {
		config := node.DefaultConfig()
		addr, transp := net.NewInmemTransport("")
		key, _ := crypto.GenerateECDSAKey()
		pubKey := fmt.Sprintf("0x%X", crypto.FromECDSAPub(&key.PublicKey))
		peer := peers.NewPeer(pubKey, addr)

		n := node.NewNode(
			config,
			peer.ID,
			key,
			participants,
			poset.NewInmemStore(participants, config.CacheSize),
			transp,
			dummy.NewInmemDummyApp(logger))

		participants.AddPeer(peer)
		nodes[key] = n
	}

	for _, n := range nodes {
		n.Init()
		n.RunAsync(false)
	}

	return nodes
}

func (n NodeList) Keys() []*ecdsa.PrivateKey {
	keys := make([]*ecdsa.PrivateKey, len(n))
	i := 0
	for key, _ := range n {
		keys[i] = key
		i++
	}
	return keys
}

func (n NodeList) Nodes() []*node.Node {
	nodes := make([]*node.Node, len(n))
	i := 0
	for _, node := range n {
		nodes[i] = node
		i++
	}
	return nodes
}

func (n NodeList) PushRandTxs(count int) {
	keys := n.Keys()
	for i := 0; i < count; i++ {
		j := rand.Intn(len(n))
		node := n[keys[j]]
		tx := []byte(fmt.Sprintf("node#%d transaction %d", node.ID(), i))
		node.PushTx(tx)
	}
}
