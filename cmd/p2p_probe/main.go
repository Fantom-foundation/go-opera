package main

import (
	"crypto/ecdsa"
	"os"
	"os/signal"
	"syscall"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/dnsdisc"
	"github.com/ethereum/go-ethereum/p2p/enode"

	"github.com/Fantom-foundation/go-opera/gossip"
	"github.com/Fantom-foundation/go-opera/opera"
)

func main() {
	backend := newProbeBackend()
	defer backend.Close()

	s := newServer(backend)
	err := s.Start()
	if err != nil {
		panic(err)
	}
	defer s.Stop()

	wait()
}

func wait() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	<-sigs
}

func newProbeBackend() *ProbeBackend {
	return &ProbeBackend{
		// mainnet
		nodeInfo: &gossip.NodeInfo{
			Network:     opera.MainNetworkID,
			Genesis:     common.HexToHash("0x4a53c5445584b3bfc20dbfb2ec18ae20037c716f3ba2d9e1da768a9deca17cb4"),
			Epoch:       197556,
			NumOfBlocks: 57715201,
		},

		quitSync: make(chan struct{}),
	}
}

func newServer(backend *ProbeBackend) *p2p.Server {
	return &p2p.Server{
		Config: p2p.Config{
			PrivateKey: anyKey(),
			Protocols:  ProbeProtocols(backend),

			DiscoveryV5: true,
			BootstrapNodesV5: []*enode.Node{
				eNode("867d5c0e27c973fbb2f9d2f9a2acd3347b92887f9cd217001a163619c07629f4b987fb0f3876c422b640d08510381565862592473fefb4591d59547bc403f4f3@54.146.98.52:47306"),
				eNode("b7041f62fa0310e5ffd0862710caaf8685e16c7143e8a7702ac2698e0673e246eb0e53d346f3acbab5f2dc439da296b39cd612207cd9067edaf6f24a29a1d1f3@116.202.109.106:5051"),
			},
		},
	}
}

func anyKey() *ecdsa.PrivateKey {
	key, err := crypto.GenerateKey()
	if err != nil {
		panic(err)
	}
	return key
}

func eNode(url string) *enode.Node {
	n, err := enode.Parse(enode.ValidSchemes, "enode://"+url)
	if err != nil {
		panic(err)
	}
	return n
}

func operaDialCandidates() enode.Iterator {
	dnsclient := dnsdisc.NewClient(dnsdisc.Config{})

	it, err := dnsclient.NewIterator()
	if err != nil {
		panic(err)
	}
	return it
}
