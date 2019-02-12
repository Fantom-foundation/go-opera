package peer_test

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/Fantom-foundation/go-lachesis/src/common/hexutil"
	"net"
	"net/rpc"
	"reflect"
	"testing"
	"time"

	"github.com/pkg/errors"

	"github.com/Fantom-foundation/go-lachesis/src/peer"
	"github.com/Fantom-foundation/go-lachesis/src/peer/fakenet"
	"github.com/Fantom-foundation/go-lachesis/src/poset"
)

var (
	expEagerSyncRequest = &peer.ForceSyncRequest{
		FromID: 0,
		Events: []poset.WireEvent{
			{
				Body: poset.WireBody{
					Transactions:         [][]byte(nil),
					SelfParentIndex:      1,
					OtherParentCreatorID: 10,
					OtherParentIndex:     0,
					CreatorID:            9,
				},
			},
		},
	}
	expEagerSyncResponse  = &peer.ForceSyncResponse{FromID: 1, Success: true}
	expFastForwardRequest = &peer.FastForwardRequest{FromID: 0}
	expSyncRequest        = &peer.SyncRequest{
		FromID: 0,
		Known:  map[uint64]int64{0: 1, 1: 2, 2: 3},
	}

	expSyncResponse = &peer.SyncResponse{
		FromID: 1,
		Events: []poset.WireEvent{
			{
				Body: poset.WireBody{
					Transactions:         [][]byte(nil),
					SelfParentIndex:      1,
					OtherParentCreatorID: 10,
					OtherParentIndex:     0,
					CreatorID:            9,
				},
			},
		},
		Known: map[uint64]int64{0: 5, 1: 5, 2: 6},
	}
	testError = errors.New("error")
)

type mockRpcClient struct {
	t    *testing.T
	err  error
	resp interface{}
}

func newRPCClient(t *testing.T, err error, resp interface{}) *mockRpcClient {
	return &mockRpcClient{t: t, err: err, resp: resp}
}

func (m *mockRpcClient) Go(serviceMethod string, args interface{},
	reply interface{}, done chan *rpc.Call) *rpc.Call {
	raw, err := json.Marshal(m.resp)
	if err != nil {
		m.t.Fatal(err)
	}
	if err := json.Unmarshal(raw, reply); err != nil {
		m.t.Fatal(err)
	}
	call := &rpc.Call{}
	call.Done = make(chan *rpc.Call, 10)
	call.Error = m.err
	call.Reply = reply
	call.Done <- call
	return call
}

func (m *mockRpcClient) Close() error {
	return m.err
}

func newClient(t *testing.T, m *mockRpcClient) *peer.Client {
	cli, err := peer.NewClient(m)
	if err != nil {
		t.Fatal(err)
	}
	return cli
}

func TestClientSync(t *testing.T) {
	ctx := context.Background()
	m := newRPCClient(t, testError, expSyncResponse)
	cli := newClient(t, m)
	defer func() {
		if err := cli.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	resp := &peer.SyncResponse{}
	if err := cli.Sync(
		ctx, expSyncRequest, resp); err != testError {
		t.Fatalf("expected error: %s, got: %s", testError, err)
	}

	m.err = nil

	if err := cli.Sync(
		ctx, expSyncRequest, resp); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(resp, expSyncResponse) {
		t.Fatalf("failed to get response, expected: %+v, got: %+v",
			expSyncResponse, resp)
	}
}

func TestClientForceSync(t *testing.T) {
	ctx := context.Background()
	m := newRPCClient(t, testError, expEagerSyncResponse)
	cli := newClient(t, m)
	defer func() {
		if err := cli.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	resp := &peer.ForceSyncResponse{}
	if err := cli.ForceSync(
		ctx, expEagerSyncRequest, resp); err != testError {
		t.Fatalf("expected error: %s, got: %s", testError, err)
	}

	m.err = nil

	if err := cli.ForceSync(
		ctx, expEagerSyncRequest, resp); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(resp, expEagerSyncResponse) {
		t.Fatalf("failed to get response, expected: %+v, got: %+v",
			expEagerSyncResponse, resp)
	}
}

func TestClientFastForward(t *testing.T) {
	expResponse := newFastForwardResponse(t)
	ctx := context.Background()
	m := newRPCClient(t, testError, expResponse)
	cli := newClient(t, m)
	defer func() {
		if err := cli.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	resp := &peer.FastForwardResponse{}
	if err := cli.FastForward(
		ctx, expFastForwardRequest, resp); err != testError {
		t.Fatalf("expected error: %s, got: %s", testError, err)
	}

	m.err = nil

	if err := cli.FastForward(
		ctx, expFastForwardRequest, resp); err != nil {
		t.Fatal(err)
	}

	checkFastForwardResponse(t, expResponse, resp)
}

func TestNewClient(t *testing.T) {
	timeout := time.Second
	conf := &peer.BackendConfig{
		ReceiveTimeout: timeout,
		ProcessTimeout: timeout,
		IdleTimeout:    timeout,
	}
	done := make(chan struct{})
	defer close(done)

	address := newAddress()
	backend := newBackend(t, conf, logger, address, done,
		expSyncResponse, 0, net.Listen)
	defer func() {
		if err := backend.Close(); err != nil {
			t.Fatal(err)
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

	resp := &peer.SyncResponse{}
	if err := cli.Sync(
		context.Background(), &peer.SyncRequest{}, resp); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(resp, expSyncResponse) {
		t.Fatalf("failed to get response, expected: %+v, got: %+v",
			expSyncResponse, resp)
	}
}

func TestFakeNet(t *testing.T) {
	timeout := time.Second
	conf := &peer.BackendConfig{
		ReceiveTimeout: timeout,
		ProcessTimeout: timeout,
		IdleTimeout:    timeout,
	}
	done := make(chan struct{})
	defer close(done)

	// Create fake network
	network := fakenet.NewNetwork()

	address := newAddress()
	backend := newBackend(t, conf, logger, address, done,
		expSyncResponse, 0, network.CreateListener)
	defer func() {
		if err := backend.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	rpcCli, err := peer.NewRPCClient(peer.TCP, address, time.Second,
		network.CreateNetConn)
	if err != nil {
		t.Fatal(err)
	}

	cli, err := peer.NewClient(rpcCli)
	if err != nil {
		t.Fatal(err)
	}

	resp := &peer.SyncResponse{}
	if err := cli.Sync(
		context.Background(), &peer.SyncRequest{}, resp); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(resp, expSyncResponse) {
		t.Fatalf("failed to get response, expected: %+v, got: %+v",
			expSyncResponse, resp)
	}
}

func newFastForwardResponse(t *testing.T) *peer.FastForwardResponse {
	frame := poset.Frame{}
	block, err := poset.NewBlockFromFrame(1, frame)
	if err != nil {
		t.Fatal(err)
	}

	return &peer.FastForwardResponse{
		FromID:   1,
		Block:    block,
		Frame:    frame,
		Snapshot: []byte("snapshot"),
	}
}

func checkFastForwardResponse(t *testing.T, exp, got *peer.FastForwardResponse) {
	if !got.Block.Equals(&exp.Block) || !got.Frame.Equals(&exp.Frame) ||
		got.FromID != exp.FromID || !bytes.Equal(got.Snapshot, exp.Snapshot) {
		t.Fatalf("bad response, expected: %+v, got: %+v", exp, got)
	}

	hash1, _ := exp.Frame.Hash()
	hash2, _ := got.Frame.Hash()
	if !bytes.Equal(hash1, hash2) {
		t.Fatalf("expected hash %s, got %s", hexutil.Encode(hash1), hexutil.Encode(hash2))

	}
}
