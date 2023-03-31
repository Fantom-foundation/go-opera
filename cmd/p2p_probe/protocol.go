package main

import (
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/core/forkid"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/discover/discfilter"
	"github.com/ethereum/go-ethereum/p2p/dnsdisc"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/enr"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/Fantom-foundation/go-opera/gossip"
	"github.com/Fantom-foundation/go-opera/opera"
)

type ProbeBackend struct {
	Progress gossip.PeerProgress
	NodeInfo *gossip.NodeInfo
	Opera    *opera.Rules
	Chain    *params.ChainConfig

	wg       sync.WaitGroup
	quitSync chan struct{}
}

func NewProbeBackend() *ProbeBackend {
	return &ProbeBackend{
		quitSync: make(chan struct{}),
	}
}

func (b *ProbeBackend) Close() {
	close(b.quitSync)
	b.wg.Wait()
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

				select {
				case <-backend.quitSync:
					return p2p.DiscQuitting
				default:
					backend.wg.Add(1)
					defer backend.wg.Done()
					return backend.handle(peer)
				}
			},
			NodeInfo: func() interface{} {
				return backend.NodeInfo
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
	log.Info("Dialed peer", "name", p.Fullname(), "ip", p.RemoteAddr().String())

	// Check useless
	useless := discfilter.Banned(p.Node().ID(), p.Node().Record())
	if !strings.Contains(strings.ToLower(p.Name()), "opera") {
		useless = true
		discfilter.Ban(p.ID())
	}
	if !p.Peer.Info().Network.Trusted && useless {
		return p2p.DiscTooManyPeers
	}

	// Execute the handshake
	if err := p.Handshake(b.NodeInfo.Network, b.Progress, b.NodeInfo.Genesis); err != nil {
		p.Log().Debug("Handshake failed", "err", err)
		if !useless {
			discfilter.Ban(p.ID())
		}
		return err
	}

	return nil
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
	info := b.NodeInfo
	e := &enrEntry{
		ForkID: forkid.NewID(b.Chain, info.Genesis, uint64(info.NumOfBlocks)),
	}

	return e
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
