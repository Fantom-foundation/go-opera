package main

import (
	"fmt"

	"github.com/ethereum/go-ethereum/core/forkid"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/dnsdisc"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/enr"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/Fantom-foundation/go-opera/gossip"
	"github.com/Fantom-foundation/go-opera/opera"
)

type ProbeBackend struct {
	nodeInfo *gossip.NodeInfo

	quitSync chan struct{}
}

func ProbeProtocols(backend *ProbeBackend) []p2p.Protocol {
	protocols := make([]p2p.Protocol, len(gossip.ProtocolVersions))
	for i, version := range gossip.ProtocolVersions {
		version := version // closure

		protocols[i] = p2p.Protocol{
			Name:    gossip.ProtocolName,
			Version: version,
			Length:  gossip.ProtocolLengths[version],
			Run: func(p *p2p.Peer, rw p2p.MsgReadWriter) error {
				peer := newPeer(version, p, rw)
				defer peer.Close()

				fmt.Printf("--> Connected to %s (%s) \n", p.Fullname(), p.RemoteAddr().String())

				select {
				case <-backend.quitSync:
					return p2p.DiscQuitting
				default:
					return backend.handle(peer)
				}
			},
			NodeInfo: func() interface{} {
				fmt.Printf("--> NodeInfo (%v) \n", backend.NodeInfo())
				return backend.NodeInfo()
			},
			PeerInfo: func(id enode.ID) interface{} {
				return nil
			},
			Attributes:     []enr.Entry{currentENREntry(backend)},
			DialCandidates: operaDialCandidates(),
		}
	}

	return protocols
}

func (b *ProbeBackend) handle(p *peer) error {
	return nil
}

func (b *ProbeBackend) Close() {
	return
}

func (b *ProbeBackend) NodeInfo() *gossip.NodeInfo {
	return b.nodeInfo
}

// ENR

// enrEntry is the ENR entry which advertises `eth` protocol on the discovery.
type enrEntry struct {
	ForkID forkid.ID // Fork identifier per EIP-2124

	// Ignore additional fields (for forward compatibility).
	Rest []rlp.RawValue `rlp:"tail"`
}

// ENRKey implements enr.Entry.
func (enrEntry) ENRKey() string {
	return "opera"
}

func currentENREntry(b *ProbeBackend) enr.Entry {
	info := b.NodeInfo()

	chainConfig := opera.MainNetRules().EvmChainConfig(
		[]opera.UpgradeHeight{
			{
				Height: 0,
				Upgrades: opera.Upgrades{
					Berlin: true,
					London: true,
					Llr:    true,
				},
			},
		})

	return &enrEntry{
		ForkID: forkid.NewID(chainConfig, info.Genesis, uint64(info.NumOfBlocks)),
	}
}

// Dial candidates

func operaDialCandidates() enode.Iterator {
	var config gossip.Config

	dnsclient := dnsdisc.NewClient(dnsdisc.Config{})

	urls := config.OperaDiscoveryURLs
	it, err := dnsclient.NewIterator(urls...)
	if err != nil {
		panic(err)
	}

	return it
}
