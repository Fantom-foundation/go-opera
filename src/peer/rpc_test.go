package peer_test

import (
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/pkg/errors"

	"github.com/Fantom-foundation/go-lachesis/src/peer"
)

type env struct {
	shutdown bool
	done     chan struct{}
	handler  *peer.Lachesis
	mtx      sync.Mutex
	err      []error
}

func newEnv(request, response interface{}, err error,
	delay, timeout time.Duration, receiver chan *peer.RPC) *env {
	done := make(chan struct{})
	handler := peer.NewLachesis(nil, receiver, timeout, timeout)

	environment := &env{
		done:    done,
		handler: handler,
	}

	// Processing simulation.
	go func() {
		for {
			var req *peer.RPC
			select {
			case <-done:
				return
			case r, ok := <-receiver:
				if !ok {
					return
				}
				if !reflect.DeepEqual(r.Command, request) {
					err := errors.Errorf("expected request %+v, got %+v",
						request, r)
					environment.mtx.Lock()
					environment.err = append(environment.err, err)
					environment.mtx.Unlock()
				}
				req = r
			}

			time.Sleep(delay)

			select {
			case <-done:
				return
			case req.RespChan <- &peer.RPCResponse{
				Response: response, Error: err}:
				close(req.RespChan)
			}
		}
	}()
	return environment
}

func (e *env) close(t *testing.T) {
	if e.shutdown {
		return
	}
	e.shutdown = true
	close(e.done)

	e.mtx.Lock()
	defer e.mtx.Unlock()
	if len(e.err) != 0 {
		t.Fatal(e.err[0])
	}
}

func TestLachesisSync(t *testing.T) {
	receiver := make(chan *peer.RPC)
	env := newEnv(expSyncRequest, expSyncResponse,
		testError, 0, time.Second, receiver)
	defer env.close(t)

	resp := &peer.SyncResponse{}
	if err := env.handler.Sync(expSyncRequest, resp); err == nil {
		t.Fatalf("expected error %s, got: error is null", testError)
	}
	env.close(t)

	receiver = make(chan *peer.RPC)
	env = newEnv(expSyncRequest, expSyncResponse,
		nil, 0, time.Second, receiver)
	defer env.close(t)

	resp = &peer.SyncResponse{}
	if err := env.handler.Sync(expSyncRequest, resp); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(resp, expSyncResponse) {
		t.Fatalf("failed to get response, expected: %+v, got: %+v",
			expSyncResponse, resp)
	}
}

func TestLachesisForceSync(t *testing.T) {
	receiver := make(chan *peer.RPC)
	env := newEnv(expEagerSyncRequest, expEagerSyncResponse,
		testError, 0, time.Second, receiver)
	defer env.close(t)

	resp := &peer.ForceSyncResponse{}
	if err := env.handler.ForceSync(expEagerSyncRequest, resp); err == nil {
		t.Fatalf("expected error %s, got: error is null", testError)
	}
	env.close(t)

	receiver = make(chan *peer.RPC)
	env = newEnv(expEagerSyncRequest, expEagerSyncResponse,
		nil, 0, time.Second, receiver)
	defer env.close(t)

	resp = &peer.ForceSyncResponse{}
	if err := env.handler.ForceSync(expEagerSyncRequest, resp); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(resp, expEagerSyncResponse) {
		t.Fatalf("failed to get response, expected: %+v, got: %+v",
			expEagerSyncResponse, resp)
	}
}

func TestLachesisFastForward(t *testing.T) {
	request := &peer.FastForwardRequest{
		FromID: 0,
	}

	expResponse := newFastForwardResponse(t)

	receiver := make(chan *peer.RPC)
	env := newEnv(request, expResponse, testError, 0, time.Second, receiver)
	defer env.close(t)

	resp := &peer.FastForwardResponse{}
	if err := env.handler.FastForward(request, resp); err == nil {
		t.Fatalf("expected error %s, got: error is null", testError)
	}
	env.close(t)

	receiver = make(chan *peer.RPC)
	env = newEnv(request, expResponse, nil, 0, time.Second, receiver)
	defer env.close(t)

	resp = &peer.FastForwardResponse{}
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
		env := newEnv(expSyncRequest, expSyncResponse,
			nil, delay/2, delay/2, nil)
		defer env.close(t)

		resp := &peer.SyncResponse{}
		if err := env.handler.Sync(
			expSyncRequest, resp); err != peer.ErrReceiverIsBusy {
			t.Fatalf("expected error %s, got: error is %v",
				peer.ErrReceiverIsBusy, err)
		}
	})

	t.Run("ProcessingTimeout", func(t *testing.T) {
		receiver := make(chan *peer.RPC)
		env := newEnv(expSyncRequest, expSyncResponse,
			nil, delay, delay/2, receiver)
		defer env.close(t)

		resp := &peer.SyncResponse{}
		if err := env.handler.Sync(
			expSyncRequest, resp); err != peer.ErrProcessingTimeout {
			t.Fatalf("expected error %s, got: error is %v",
				peer.ErrProcessingTimeout, err)
		}
	})
}
