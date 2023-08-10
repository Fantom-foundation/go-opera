// Copyright 2020 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package gossip

import (
	"github.com/ethereum/go-ethereum/eth/protocols/snap"
	"github.com/ethereum/go-ethereum/p2p/enode"
)

// snapHandler implements the snap.Backend interface to handle the various network
// packets that are sent as replies or broadcasts.
type snapHandler handler

func (h *snapHandler) Chain() snap.BlockChain {
	return h.chain
}

// RunPeer is invoked when a peer joins on the `snap` protocol.
func (h *snapHandler) RunPeer(peer *snap.Peer, hand snap.Handler) error {
	return (*handler)(h).runSnapExtension(peer, hand)
}

// PeerInfo retrieves all known `snap` information about a peer.
func (h *snapHandler) PeerInfo(id enode.ID) interface{} {
	if p := h.peers.Peer(id.String()); p != nil {
		if p.snapExt != nil {
			return p.snapExt.info()
		}
	}
	return nil
}

// Handle is invoked from a peer's message handler when it receives a new remote
// message that the handler couldn't consume and serve itself.
func (h *snapHandler) Handle(peer *snap.Peer, packet snap.Packet) error {
	return h.snapLeecher.DeliverSnapPacket(peer, packet)
}

// runSnapExtension registers a `snap` peer into the joint eth/snap peerset and
// starts handling inbound messages. As `snap` is only a satellite protocol to
// `eth`, all subsystem registrations and lifecycle management will be done by
// the main `eth` handler to prevent strange races.
func (h *handler) runSnapExtension(peer *snap.Peer, handler snap.Handler) error {
	h.peerWG.Add(1)
	defer h.peerWG.Done()
	if err := h.peers.RegisterSnapExtension(peer); err != nil {
		logger := peer.Log().Error
		if err == errSnapWithoutOpera {
			logger = peer.Log().Trace
		}
		logger("Snapshot extension registration failed", "err", err)
		return err
	}
	return handler(peer)
}
