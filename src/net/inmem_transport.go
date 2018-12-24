package net

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/rs/xid"
)

var (
	inmemMedium     = make(map[string]*InmemTransport)
	inmemMediumSync sync.RWMutex
)

// NewInmemAddr returns a new in-memory addr with
// a randomly generate UUID as the ID.
func NewInmemAddr() string {
	return xid.New().String()
}

// InmemTransport implements the Transport interface, to allow lachesis to be
// tested in-memory without going over a network.
type InmemTransport struct {
	consumerCh chan RPC
	localAddr  string
	timeout    time.Duration
}

// NewInmemTransport is used to initialize a new transport
// and generates a random local address if none is specified
func NewInmemTransport(addr string) (string, *InmemTransport) {
	if addr == "" {
		addr = NewInmemAddr()
	}
	trans := &InmemTransport{
		consumerCh: make(chan RPC, 16),
		localAddr:  addr,
		timeout:    50 * time.Millisecond,
	}

	inmemMediumSync.Lock()
	inmemMedium[addr] = trans
	inmemMediumSync.Unlock()

	return addr, trans
}

// Consumer implements the Transport interface.
func (i *InmemTransport) Consumer() <-chan RPC {
	return i.consumerCh
}

// LocalAddr implements the Transport interface.
func (i *InmemTransport) LocalAddr() string {
	return i.localAddr
}

// Sync implements the Transport interface.
func (i *InmemTransport) Sync(target string, args *SyncRequest, resp *SyncResponse) error {
	return i.makeRPC(target, args, resp, nil, i.timeout)
}

// EagerSync implements the Transport interface.
func (i *InmemTransport) EagerSync(target string, args *EagerSyncRequest, resp *EagerSyncResponse) error {
	return i.makeRPC(target, args, resp, nil, i.timeout)
}

// FastForward implements the Transport interface.
func (i *InmemTransport) FastForward(target string, args *FastForwardRequest, resp *FastForwardResponse) error {
	return i.makeRPC(target, args, resp, nil, i.timeout)
}

func (i *InmemTransport) makeRPC(target string, args, resp interface{}, r io.Reader, timeout time.Duration) (err error) {
	args, err = deepRequestCopy(args)
	if err != nil {
		return
	}

	inmemMediumSync.RLock()
	peer, ok := inmemMedium[target]
	inmemMediumSync.RUnlock()
	if !ok {
		err = fmt.Errorf("failed to connect to peer: %v", target)
		return
	}

	// Send the RPC over
	respCh := make(chan RPCResponse)
	peer.consumerCh <- RPC{
		Command:  args,
		Reader:   r,
		RespChan: respCh,
	}

	// Wait for a response
	select {
	case rpcResp := <-respCh:
		err = deepResponceCopy(rpcResp.Response, resp)
		if err == nil && rpcResp.Error != nil {
			err = rpcResp.Error
		}
	case <-time.After(timeout):
		err = fmt.Errorf("command timed out")
	}
	return
}

// Close is used to permanently disable the transport
func (i *InmemTransport) Close() {
	inmemMediumSync.Lock()
	delete(inmemMedium, i.localAddr)
	inmemMediumSync.Unlock()
}

func deepRequestCopy(src interface{}) (dst interface{}, err error) {
	data, err := json.Marshal(src)
	if err != nil {
		return
	}

	switch t := src.(type) {
	case *SyncRequest:
		dst = new(SyncRequest)
	case *EagerSyncRequest:
		dst = new(EagerSyncRequest)
	case *FastForwardRequest:
		dst = new(FastForwardRequest)
	default:
		err = fmt.Errorf("Unknown request type %s", t)
		return
	}

	err = json.Unmarshal(data, dst)
	return
}

func deepResponceCopy(src, dst interface{}) error {
	data, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dst)
}
