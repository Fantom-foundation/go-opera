package node

import (
	"crypto/ecdsa"
	"fmt"
	"math/rand"
	_ "testing"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/andrecronje/lachesis/src/crypto"
	"github.com/andrecronje/lachesis/src/dummy"
	"github.com/andrecronje/lachesis/src/net"
	"github.com/andrecronje/lachesis/src/peers"
	"github.com/andrecronje/lachesis/src/poset"
)

const delay = 100 * time.Millisecond

// NodeList is a list of connected nodes for tests purposes
type NodeList map[*ecdsa.PrivateKey]*Node

// NewNodeList makes, fills and runs NodeList instance
func NewNodeList(count int, logger *logrus.Logger) NodeList {
	nodes := make(NodeList, count)
	participants := peers.NewPeers()

	for i := 0; i < count; i++ {
		config := DefaultConfig()
		addr, transp := net.NewInmemTransport("")
		key, _ := crypto.GenerateECDSAKey()
		pubKey := fmt.Sprintf("0x%X", crypto.FromECDSAPub(&key.PublicKey))
		peer := peers.NewPeer(pubKey, addr)

		n := NewNode(
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
		n.RunAsync(true)
	}

	return nodes
}

// Keys returns the all PrivateKeys slice
func (n NodeList) Keys() []*ecdsa.PrivateKey {
	keys := make([]*ecdsa.PrivateKey, len(n))
	i := 0
	for key, _ := range n {
		keys[i] = key
		i++
	}
	return keys
}

// Values returns the all nodes slice
func (n NodeList) Values() []*Node {
	nodes := make([]*Node, len(n))
	i := 0
	for _, node := range n {
		nodes[i] = node
		i++
	}
	return nodes
}

// StartRandTxStream sends random txs to nodes until stop() called
func (n NodeList) StartRandTxStream() (stop func()) {
	stopCh := make(chan struct{})

	stop = func() {
		close(stopCh)
	}

	go func() {
		seq := 0
		for {
			select {
			case <-stopCh:
				return
			case <-time.After(delay):
				keys := n.Keys()
				count := len(n)
				for i := 0; i < count; i++ {
					j := rand.Intn(count)
					node := n[keys[j]]
					tx := []byte(fmt.Sprintf("node#%d transaction %d", node.ID(), seq))
					node.PushTx(tx)
					seq++
				}
			}
		}
	}()

	return
}

// WaitForBlock waits until the target block has retrieved a state hash from the app
func (n NodeList) WaitForBlock(target int64) {
LOOP:
	for {
		time.Sleep(delay)
		for _, node := range n {
			if target > node.GetLastBlockIndex() {
				continue LOOP
			}
			block, _ := node.GetBlock(target)
			if len(block.StateHash()) == 0 {
				continue LOOP
			}
		}
		return
	}
}
