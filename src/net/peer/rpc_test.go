package peer_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/Fantom-foundation/go-lachesis/src/net"
	"github.com/Fantom-foundation/go-lachesis/src/net/peer"
)

type env struct {
	shutdown bool
	done     chan struct{}
	handler  *peer.Lachesis
}

func newEnv(t *testing.T, request, response interface{}, err error,
	delay, timeout time.Duration, receiver chan *net.RPC) *env {
	done := make(chan struct{})
	handler := peer.NewLachesis(nil, receiver, timeout, timeout)

	// Processing simulation.
	go func() {
		for {
			var req *net.RPC
			select {
			case <-done:
				return
			case r, ok := <-receiver:
				if !ok {
					return
				}
				if !reflect.DeepEqual(r.Command, request) {
					t.Fatalf("expected request %+v, got %+v",
						request, r)
				}
				req = r
			}

			time.Sleep(delay)

			select {
			case <-done:
				return
			case req.RespChan <- &net.RPCResponse{
				Response: response, Error: err}:
				close(req.RespChan)
			}
		}
	}()
	return &env{
		done:    done,
		handler: handler,
	}
}

func (e *env) Close() {
	if e.shutdown {
		return
	}
	e.shutdown = true
	close(e.done)
}

func TestLachesisSync(t *testing.T) {
	receiver := make(chan *net.RPC)
	env := newEnv(t, expSyncRequest, expSyncResponse,
		testError, 0, time.Second, receiver)
	defer env.Close()

	resp := &net.SyncResponse{}
	if err := env.handler.Sync(expSyncRequest, resp); err == nil {
		t.Fatalf("expected error %s, got: error is null", testError)
	}
	env.Close()

	receiver = make(chan *net.RPC)
	env = newEnv(t, expSyncRequest, expSyncResponse,
		nil, 0, time.Second, receiver)
	defer env.Close()

	resp = &net.SyncResponse{}
	if err := env.handler.Sync(expSyncRequest, resp); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(resp, expSyncResponse) {
		t.Fatalf("failed to get response, expected: %+v, got: %+v",
			expSyncResponse, resp)
	}
}

func TestLachesisForceSync(t *testing.T) {
	receiver := make(chan *net.RPC)
	env := newEnv(t, expEagerSyncRequest, expEagerSyncResponse,
		testError, 0, time.Second, receiver)
	defer env.Close()

	resp := &net.EagerSyncResponse{}
	if err := env.handler.ForceSync(expEagerSyncRequest, resp); err == nil {
		t.Fatalf("expected error %s, got: error is null", testError)
	}
	env.Close()

	receiver = make(chan *net.RPC)
	env = newEnv(t, expEagerSyncRequest, expEagerSyncResponse,
		nil, 0, time.Second, receiver)
	defer env.Close()

	resp = &net.EagerSyncResponse{}
	if err := env.handler.ForceSync(expEagerSyncRequest, resp); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(resp, expEagerSyncResponse) {
		t.Fatalf("failed to get response, expected: %+v, got: %+v",
			expEagerSyncResponse, resp)
	}
}

func TestLachesisFastForward(t *testing.T) {
	request := &net.FastForwardRequest{
		FromID: 0,
	}

	expResponse := newFastForwardResponse(t)

	receiver := make(chan *net.RPC)
	env := newEnv(t, request, expResponse, testError, 0, time.Second, receiver)
	defer env.Close()

	resp := &net.FastForwardResponse{}
	if err := env.handler.FastForward(request, resp); err == nil {
		t.Fatalf("expected error %s, got: error is null", testError)
	}
	env.Close()

	receiver = make(chan *net.RPC)
	env = newEnv(t, request, expResponse, nil, 0, time.Second, receiver)
	defer env.Close()

	resp = &net.FastForwardResponse{}
	if err := env.handler.FastForward(request, resp); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(resp, expResponse) {
		t.Fatalf("failed to get response, expected: %+v, got: %+v",
			expResponse, resp)
	}
}

func TestTimeout(t *testing.T) {
	delay := time.Second

	t.Run("ReceiverIsBusy", func(t *testing.T) {
		env := newEnv(t, expSyncRequest, expSyncResponse,
			nil, delay/2, delay/2, nil)
		defer env.Close()

		resp := &net.SyncResponse{}
		if err := env.handler.Sync(
			expSyncRequest, resp); err != peer.ErrReceiverIsBusy {
			t.Fatalf("expected error %s, got: error is %v",
				peer.ErrReceiverIsBusy, err)
		}
	})

	t.Run("ProcessingTimeout", func(t *testing.T) {
		receiver := make(chan *net.RPC)
		env := newEnv(t, expSyncRequest, expSyncResponse,
			nil, delay, delay/2, receiver)
		defer env.Close()

		resp := &net.SyncResponse{}
		if err := env.handler.Sync(
			expSyncRequest, resp); err != peer.ErrProcessingTimeout {
			t.Fatalf("expected error %s, got: error is %v",
				peer.ErrProcessingTimeout, err)
		}
	})
}
