package peer_test

import (
	"context"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/Fantom-foundation/go-lachesis/src/net"
	"github.com/Fantom-foundation/go-lachesis/src/net/peer"
)

var logger logrus.FieldLogger

type node struct {
	address   string
	transport *peer.Peer
}

func TestPeerClient(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	limit := 2
	target := "1:2"
	timeout := time.Second

	t.Run("Sync", func(t *testing.T) {
		createFu := func(target string,
			timeout time.Duration) (peer.SyncClient, error) {
			return peer.NewClient(
				newRPCClient(t, nil, expSyncResponse))
		}

		producer := peer.NewProducer(limit, timeout, createFu)
		tr := peer.NewTransport(logger, producer, nil)
		defer func() {
			if err := tr.Close(); err != nil {
				t.Fatal(err)
			}
		}()

		resp := &net.SyncResponse{}
		if err := tr.Sync(
			ctx, target, expSyncRequest, resp); err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(resp, expSyncResponse) {
			t.Fatalf("failed to get response, expected: %+v, got: %+v",
				expSyncResponse, resp)
		}
	})

	t.Run("ForceSync", func(t *testing.T) {
		createFu := func(target string,
			timeout time.Duration) (peer.SyncClient, error) {
			return peer.NewClient(
				newRPCClient(t, nil, expEagerSyncResponse))
		}

		producer := peer.NewProducer(limit, timeout, createFu)
		tr := peer.NewTransport(logger, producer, nil)
		defer func() {
			if err := tr.Close(); err != nil {
				t.Fatal(err)
			}
		}()

		resp := &net.EagerSyncResponse{}
		if err := tr.ForceSync(
			ctx, target, expEagerSyncRequest, resp); err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(resp, expEagerSyncResponse) {
			t.Fatalf("failed to get response, expected: %+v, got: %+v",
				expEagerSyncResponse, resp)
		}
	})

	t.Run("FastForward", func(t *testing.T) {
		expResponse := newFastForwardResponse(t)

		createFu := func(target string,
			timeout time.Duration) (peer.SyncClient, error) {
			return peer.NewClient(
				newRPCClient(t, nil, expResponse))
		}

		producer := peer.NewProducer(limit, timeout, createFu)
		tr := peer.NewTransport(logger, producer, nil)
		defer func() {
			if err := tr.Close(); err != nil {
				t.Fatal(err)
			}
		}()

		resp := &net.FastForwardResponse{}
		if err := tr.FastForward(
			ctx, target, expFastForwardRequest, resp); err != nil {
			t.Fatal(err)
		}

		checkFastForwardResponse(t, expResponse, resp)
	})
}

func TestPeerClose(t *testing.T) {
	connectLimit := 2
	netSize := 2
	timeout := time.Second
	reqNum := 10
	done := make(chan struct{})
	defer close(done)

	network := newNetwork(t, done, logger, connectLimit,
		expSyncResponse, timeout, timeout, timeout, netSize)
	defer networkStop(t, network)

	runClient := func(cli, srv *node, errorIsNil bool) {
		req := expSyncRequest
		resp := &net.SyncResponse{}
		err := cli.transport.Sync(context.Background(),
			srv.address, req, resp)
		if errorIsNil && err != nil {
			t.Fatalf("exptected error: nil, got: %v", err)
		} else if !errorIsNil && err == nil {
			t.Fatalf("exptected error: not nil, got: %v", err)
		}
	}

	// Test normal communication.
	for i := 0; i < reqNum; i++ {
		runClient(network[0], network[1], true)

	}

	// Test break connection.
	if err := network[1].transport.Close(); err != nil {
		t.Fatal(err)
	}

	// Test after shutdown.
	for i := 0; i < reqNum; i++ {
		runClient(network[0], network[1], false)

	}
}

func newNode(t *testing.T, done chan struct{}, logger logrus.FieldLogger,
	limit int, resp interface{}, backendTimeout, clientTimeout,
	idleTimeout time.Duration) *node {
	createFu := func(target string,
		timeout time.Duration) (peer.SyncClient, error) {
		rClient, err := peer.NewRPCClient(peer.TCP, target, timeout)
		if err != nil {
			return nil, err
		}
		return peer.NewClient(rClient)
	}
	producer := peer.NewProducer(limit, clientTimeout, createFu)
	address := newAddress()
	backend := newBackend(t, logger, address, done, resp,
		backendTimeout, backendTimeout, idleTimeout, 0)
	return &node{address, peer.NewTransport(logger, producer, backend)}
}

func newNetwork(t *testing.T, done chan struct{}, logger logrus.FieldLogger,
	limit int, resp interface{}, backendTimeout, dialTimeout,
	idleTimeout time.Duration, size int) (network []*node) {
	for i := 0; i < size; i++ {
		network = append(network, newNode(t, done, logger, limit,
			resp, backendTimeout, dialTimeout, idleTimeout))
	}
	return network
}

func networkStop(t *testing.T, network []*node) {
	for k := range network {
		if err := network[k].transport.Close(); err != nil {
			t.Fatal(err)
		}
	}
}

func newTestLogger() *logrus.Logger {
	logger := logrus.New()
	logger.Level = logrus.FatalLevel
	return logger
}

func TestMain(m *testing.M) {
	logger = newTestLogger()

	os.Exit(m.Run())
}
