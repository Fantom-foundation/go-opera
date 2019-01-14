package peer_test

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/Fantom-foundation/go-lachesis/src/net"
	"github.com/Fantom-foundation/go-lachesis/src/net/peer"
	"github.com/Fantom-foundation/go-lachesis/src/utils"
)

func newAddress() string {
	return "localhost:" + strconv.Itoa(int(utils.FreePort(peer.TCP)))
}

func newBackend(t *testing.T, logger logrus.FieldLogger, address string,
	done chan struct{}, resp interface{}, receiveTimeout,
	processTimeout, idleTimeout, delay time.Duration) *peer.Backend {
	backend := peer.NewBackend(
		receiveTimeout, processTimeout, idleTimeout, logger)
	receiver := backend.ReceiverChannel()

	go func() {
		for {
			select {
			case <-done:
				return
			case req := <-receiver:
				// Delay response.
				time.Sleep(delay)

				req.RespChan <- &net.RPCResponse{
					Response: resp,
				}
			}
		}
	}()

	if err := backend.ListenAndServe(peer.TCP, address); err != nil {
		t.Fatal(err)
	}

	return backend
}

func TestBackendClose(t *testing.T) {
	srvTimeout := time.Second * 30

	done := make(chan struct{})
	defer close(done)

	reqNumber := 1000
	result := make(chan error, reqNumber)
	defer close(result)

	address := newAddress()
	backend := newBackend(t, logger, address, done, expSyncResponse,
		srvTimeout, srvTimeout, srvTimeout, srvTimeout)
	defer backend.Close()

	rpcCli, err := peer.NewRPCClient(peer.TCP, address, time.Second)
	if err != nil {
		t.Fatal(err)
	}

	cli, err := peer.NewClient(rpcCli)
	if err != nil {
		t.Fatal(err)
	}

	request := func() {
		resp := &net.SyncResponse{}
		result <- cli.Sync(context.Background(), &net.SyncRequest{}, resp)
	}

	for i := 0; i < reqNumber; i++ {
		go request()
	}

	if err := backend.Close(); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < reqNumber; i++ {
		err := <-result
		if err == nil {
			t.Fatal("error must be not nil, got: nil")
		}
	}
}
