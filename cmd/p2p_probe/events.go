package main

import (
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/p2p"

	"github.com/Fantom-foundation/go-opera/gossip"
	"github.com/Fantom-foundation/go-opera/gossip/protocols/dag/dagstream"
	"github.com/Fantom-foundation/go-opera/gossip/protocols/dag/dagstream/dagstreamleecher"
	"github.com/Fantom-foundation/go-opera/inter"
)

type dagChunk struct {
	SessionID uint32
	Done      bool
	IDs       hash.Events
	Events    inter.EventPayloads
}

func (p *peer) RequestEventsStream(r dagstream.Request) error {
	p.Log().Info("PEER RequestEvents", "request", r)
	return p2p.Send(p.rw, gossip.RequestEventsStream, r)
}

func (b *ProbeBackend) startDagLeecher(epoch idx.Epoch) *dagstreamleecher.Leecher {
	leecher := dagstreamleecher.New(epoch, true, dagstreamleecher.DefaultConfig(), dagstreamleecher.Callbacks{
		IsProcessed: func(hash.Event) bool { return false },
		RequestChunk: func(peer string, r dagstream.Request) error {
			p := b.peers.Peer(peer)
			if p == nil {
				return errPeerNotRegistered
			}
			return p.RequestEventsStream(r)
		},
		Suspend: func(_ string) bool {
			return false
		},
		PeerEpoch: func(peer string) idx.Epoch {
			p := b.peers.Peer(peer)
			if p == nil {
				return 0
			}
			return p.progress.Epoch
		},
	})

	return leecher
}
