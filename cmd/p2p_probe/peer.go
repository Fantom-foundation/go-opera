package main

import (
	"github.com/ethereum/go-ethereum/p2p"
)

type peer struct {
	*p2p.Peer
	version uint // Protocol version negotiated
	id      string
	rw      p2p.MsgReadWriter
}

func newPeer(version uint, p *p2p.Peer, rw p2p.MsgReadWriter) *peer {
	peer := &peer{
		Peer:    p,
		version: version,
		id:      p.ID().String(),
		rw:      rw,
	}

	return peer
}

func (p *peer) Close() {
}
