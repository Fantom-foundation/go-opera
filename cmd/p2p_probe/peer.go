package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/p2p"

	"github.com/Fantom-foundation/go-opera/gossip"
	"github.com/Fantom-foundation/go-opera/inter"
)

const (
	handshakeTimeout   = 5 * time.Second
	protocolMaxMsgSize = inter.ProtocolMaxMsgSize // Maximum cap on the size of a protocol message

	softResponseLimitSize = 2 * 1024 * 1024    // Target maximum size of returned events, or other data.
	softLimitItems        = 250                // Target maximum number of events or transactions to request/response
	hardLimitItems        = softLimitItems * 4 // Maximum number of events or transactions to request/response
)

type errCode int

type peer struct {
	*p2p.Peer
	version uint // Protocol version negotiated

	progress gossip.PeerProgress

	rw p2p.MsgReadWriter
	sync.RWMutex
}

func newPeer(version uint, p *p2p.Peer, rw p2p.MsgReadWriter) *peer {
	peer := &peer{
		Peer:    p,
		version: version,
		rw:      rw,
	}

	return peer
}

func (p *peer) Close() {
}

// handshakeData is the network packet for the initial handshake message
type handshakeData struct {
	ProtocolVersion uint32
	NetworkID       uint64
	Genesis         common.Hash
}

// Handshake executes the protocol handshake, negotiating version number,
// network IDs, difficulties, head and genesis object.
func (p *peer) Handshake(network uint64, progress gossip.PeerProgress, genesis common.Hash) error {
	// Send out own handshake in a new thread
	errc := make(chan error, 2)
	var handshake handshakeData // safe to read after two values have been received from errc

	go func() {
		// send both HandshakeMsg and ProgressMsg
		err := p2p.Send(p.rw, gossip.HandshakeMsg, &handshakeData{
			ProtocolVersion: uint32(p.version),
			NetworkID:       0, // TODO: set to `network` after all nodes updated to #184
			Genesis:         genesis,
		})
		if err != nil {
			errc <- err
		}
		errc <- p2p.Send(p.rw, gossip.ProgressMsg, progress)
	}()
	go func() {
		errc <- p.readStatus(network, &handshake, genesis)
		// do not expect ProgressMsg here, because eth62 clients won't send it
	}()
	timeout := time.NewTimer(handshakeTimeout)
	defer timeout.Stop()
	for i := 0; i < 2; i++ {
		select {
		case err := <-errc:
			if err != nil {
				return err
			}
		case <-timeout.C:
			return p2p.DiscReadTimeout
		}
	}
	return nil
}

func (p *peer) readStatus(network uint64, handshake *handshakeData, genesis common.Hash) (err error) {
	msg, err := p.rw.ReadMsg()
	if err != nil {
		return err
	}
	if msg.Code != gossip.HandshakeMsg {
		return errResp(gossip.ErrNoStatusMsg, "first msg has code %x (!= %x)", msg.Code, gossip.HandshakeMsg)
	}
	if msg.Size > protocolMaxMsgSize {
		return errResp(gossip.ErrMsgTooLarge, "%v > %v", msg.Size, protocolMaxMsgSize)
	}
	// Decode the handshake and make sure everything matches
	if err := msg.Decode(&handshake); err != nil {
		return errResp(gossip.ErrDecode, "msg %v: %v", msg, err)
	}

	// TODO: rm after all the nodes updated to #184
	if handshake.NetworkID == 0 {
		handshake.NetworkID = network
	}

	if handshake.Genesis != genesis {
		return errResp(gossip.ErrGenesisMismatch, "%x (!= %x)", handshake.Genesis[:8], genesis[:8])
	}
	if handshake.NetworkID != network {
		return errResp(gossip.ErrNetworkIDMismatch, "%d (!= %d)", handshake.NetworkID, network)
	}
	if uint(handshake.ProtocolVersion) != p.version {
		return errResp(gossip.ErrProtocolVersionMismatch, "%d (!= %d)", handshake.ProtocolVersion, p.version)
	}
	return nil
}

func (p *peer) SetProgress(x gossip.PeerProgress) {
	p.Lock()
	defer p.Unlock()

	p.progress = x

	p.Log().Info("PEER PROGRESS", "atropos", x.LastBlockAtropos, "name", p.Fullname(), "addr", p.RemoteAddr().String())
}

func errResp(code errCode, format string, v ...interface{}) error {
	return fmt.Errorf("%v - %v", code, fmt.Sprintf(format, v...))
}
