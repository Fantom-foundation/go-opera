package main

import (
	"fmt"

	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/enr"

	"github.com/Fantom-foundation/go-opera/gossip"
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

				fmt.Printf("Connected to %s (%s) \n", p.Fullname(), p.RemoteAddr().String())

				select {
				case <-backend.quitSync:
					return p2p.DiscQuitting
				default:
					return backend.handle(peer)
				}
			},
			NodeInfo: func() interface{} {
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
type enrEntry struct {
	backend *ProbeBackend
}

func (e *enrEntry) ENRKey() string {
	return ""
}

func currentENREntry(b *ProbeBackend) enr.Entry {
	return &enrEntry{
		backend: b,
	}
}
