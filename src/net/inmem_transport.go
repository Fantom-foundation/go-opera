package net

import (
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
	rpcResp, err := i.makeRPC(target, args, nil, i.timeout)
	if err != nil {
		return err
	}

	// Copy the result back
	out := rpcResp.Response.(*SyncResponse)
	*resp = *out
	return nil
}

// Sync implements the Transport interface.
func (i *InmemTransport) EagerSync(target string, args *EagerSyncRequest, resp *EagerSyncResponse) error {
	rpcResp, err := i.makeRPC(target, args, nil, i.timeout)
	if err != nil {
		return err
	}

	// Copy the result back
	out := rpcResp.Response.(*EagerSyncResponse)
	*resp = *out
	return nil
}

// FastForward implements the Transport interface.
func (i *InmemTransport) FastForward(target string, args *FastForwardRequest, resp *FastForwardResponse) error {
	rpcResp, err := i.makeRPC(target, args, nil, i.timeout)
	if err != nil {
		return err
	}

	// Copy the result back
	out := rpcResp.Response.(*FastForwardResponse)
	*resp = *out
	return nil
}

func (i *InmemTransport) makeRPC(target string, args interface{}, r io.Reader, timeout time.Duration) (rpcResp RPCResponse, err error) {
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
	case rpcResp = <-respCh:
		if rpcResp.Error != nil {
			err = rpcResp.Error
		}
	case <-time.After(timeout):
		err = fmt.Errorf("command timed out")
	}
	return
}

// Close is used to permanently disable the transport
func (i *InmemTransport) Close() error {
	inmemMediumSync.Lock()
	delete(inmemMedium, i.localAddr)
	inmemMediumSync.Unlock()
	return nil
}
