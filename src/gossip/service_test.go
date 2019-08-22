package gossip

import (
	"crypto/rand"

	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
)

func getProtocol(svc node.Service, name string, version uint) *p2p.Protocol {
	for _, p := range svc.Protocols() {
		if p.Name == name && p.Version == version {
			return &p
		}
	}
	return nil
}

func runProtocol(name string, p *p2p.Protocol, errc chan<- error) *p2p.MsgPipeRW {
	// Generate a random id and create the peer
	var id enode.ID
	rand.Read(id[:])
	peer := p2p.NewPeer(id, name, nil)

	app, net := p2p.MsgPipe()

	go func() {
		err := p.Run(peer, app)
		if err != nil {
			errc <- err
		}
	}()

	return net
}
