package node

import (
	"crypto/ecdsa"
	"fmt"
	"math/rand"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	"github.com/Fantom-foundation/go-lachesis/src/dummy"
	"github.com/Fantom-foundation/go-lachesis/src/peer"
	"github.com/Fantom-foundation/go-lachesis/src/peer/fakenet"
	"github.com/Fantom-foundation/go-lachesis/src/peers"
	"github.com/Fantom-foundation/go-lachesis/src/poset"
)

const delay = 100 * time.Millisecond

// NodeList is a list of connected nodes for tests purposes
type NodeList map[*ecdsa.PrivateKey]*Node

// NewNodeList makes, fills and runs NodeList instance
func NewNodeList(count int, logger *logrus.Logger) NodeList {
	config := DefaultConfig()
	syncBackConfig := peer.NewBackendConfig()

	config.Logger = logger
	network := fakenet.NewNetwork()
	createFu := func(target string,
		timeout time.Duration) (peer.SyncClient, error) {
		rpcCli, err := peer.NewRPCClient(
			peer.TCP, target, time.Second, network.CreateNetConn)
		if err != nil {
			return nil, err
		}

		return peer.NewClient(rpcCli)
	}

	participants := peers.NewPeers()
	keys := make(map[*peers.Peer]*ecdsa.PrivateKey)
	for i := 0; i < count; i++ {
		addr := network.RandomAddress()
		key, _ := crypto.GenerateECDSAKey()
		pubKey := fmt.Sprintf("0x%X", crypto.FromECDSAPub(&key.PublicKey))
		peer2 := peers.NewPeer(pubKey, addr)
		participants.AddPeer(peer2)
		keys[peer2] = key
	}

	nodes := make(NodeList, count)
	for _, peer2 := range participants.ToPeerSlice() {
		key := keys[peer2]

		producer := peer.NewProducer(config.CacheSize, time.Second, createFu)
		backend := peer.NewBackend(
			syncBackConfig, logger, network.CreateListener)
		if err := backend.ListenAndServe(peer.TCP, peer2.NetAddr); err != nil {
			logger.Panic(err)
		}
		transport := peer.NewTransport(logger, producer, backend)
		n := NewNode(
			config,
			peer2.ID,
			key,
			participants,
			poset.NewInmemStore(participants, config.CacheSize, nil),
			transport,
			dummy.NewInmemDummyApp(logger))
		if err := n.Init(); err != nil {
			logger.Fatal(err)
		}
		n.RunAsync(true)
		nodes[key] = n
	}

	return nodes
}

// Keys returns the all PrivateKeys slice
func (n NodeList) Keys() []*ecdsa.PrivateKey {
	keys := make([]*ecdsa.PrivateKey, len(n))
	i := 0
	for key := range n {
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
					if err := node.PushTx(tx); err != nil {
						panic(err)
					}
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
			if len(block.GetStateHash()) == 0 {
				continue LOOP
			}
		}
		return
	}
}
