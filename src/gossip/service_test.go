package gossip

import (
	"crypto/rand"
	"fmt"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/stretchr/testify/assert"
)

func TestServicePing(t *testing.T) {
	assertar := assert.New(t)

	db := NewMemStore()

	svc, err := NewService(&DefaultConfig, new(event.TypeMux), db, nil)
	assertar.NoError(err)

	tests := []struct {
		reqCode     uint64
		reqBody     interface{}
		protocolErr error
		respCode    uint64
		respBody    interface{}
	}{
		{
			reqCode:     PongMsg,
			reqBody:     "1111",
			protocolErr: errResp(ErrInvalidMsgCode, "ping expected"),
		},
		{
			reqCode:     PingMsg,
			reqBody:     "ping",
			protocolErr: nil,
			respCode:    PongMsg,
			respBody:    "Hello, ping!",
		},
	}

	protocol := getProtocol(svc, "ping", 1)

	for i, test := range tests {
		step := fmt.Sprintf("test-%d", i)

		errc := make(chan error, 1)
		net := runProtocol(step, protocol, errc)

		expect := test
		go func() {
			err := p2p.ExpectMsg(net, expect.respCode, expect.respBody)
			errc <- err
		}()

		err := p2p.Send(net, test.reqCode, test.reqBody)
		if err != nil {
			t.Fatal(err)
		}

		select {
		case err := <-errc:
			assertar.Equal(test.protocolErr, err, step)
		case <-time.After(2 * time.Second):
			t.Errorf("protocol did not shut down within 2 seconds")
		}

		net.Close()
	}

}

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
