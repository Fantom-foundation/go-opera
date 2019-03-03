package peer_test

import (
	"context"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/Fantom-foundation/go-lachesis/src/peer"
	"github.com/Fantom-foundation/go-lachesis/src/utils"
)

func newAddress() string {
	return "localhost:" + strconv.Itoa(int(utils.FreePort(peer.TCP)))
}

func newBackend(t *testing.T, conf *peer.BackendConfig,
	logger logrus.FieldLogger, address string, done chan struct{},
	resp interface{}, delay time.Duration,
	listenerFunc peer.CreateListenerFunc) *peer.Backend {
	backend := peer.NewBackend(conf, logger, listenerFunc)
	receiver := backend.ReceiverChannel()

	go func() {
		for {
			select {
			case <-done:
				return
			case req := <-receiver:
				// Delay response.
				time.Sleep(delay)

				req.RespChan <- &peer.RPCResponse{
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
	conf := &peer.BackendConfig{
		ReceiveTimeout: srvTimeout,
		ProcessTimeout: srvTimeout,
		IdleTimeout:    srvTimeout,
	}

	done := make(chan struct{})
	defer close(done)

	reqNumber := 1000
	result := make(chan error, reqNumber)
	defer close(result)

	address := newAddress()
	backend := newBackend(t, conf, logger, address, done,
		expSyncResponse, srvTimeout, net.Listen)
	defer func() {
		if err := backend.Close(); err != nil {
			panic(err)
		}
	}()

	rpcCli, err := peer.NewRPCClient(
		peer.TCP, address, time.Second, net.DialTimeout)
	if err != nil {
		t.Fatal(err)
	}

	cli, err := peer.NewClient(rpcCli)
	if err != nil {
		t.Fatal(err)
	}

	request := func() {
		resp := &peer.SyncResponse{}
		result <- cli.Sync(context.Background(), &peer.SyncRequest{}, resp)
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
