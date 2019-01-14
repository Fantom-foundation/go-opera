package peer

import (
	"time"

	"github.com/Fantom-foundation/go-lachesis/src/net"
)

// RPC Methods.
const (
	MethodSync        = "Lachesis.Sync"
	MethodForceSync   = "Lachesis.ForceSync"
	MethodFastForward = "Lachesis.FastForward"
)

// Lachesis implements Lachesis synchronization methods.
type Lachesis struct {
	done           chan struct{}
	receiver       chan *net.RPC
	processTimeout time.Duration
	receiveTimeout time.Duration
}

// NewLachesis creates new Lachesis RPC handler.
func NewLachesis(done chan struct{}, receiver chan *net.RPC,
	receiveTimeout, processTimeout time.Duration) *Lachesis {
	return &Lachesis{
		done:           done,
		receiver:       receiver,
		processTimeout: processTimeout,
		receiveTimeout: receiveTimeout,
	}
}

// Sync handles sync requests.
func (r *Lachesis) Sync(
	req *net.SyncRequest, resp *net.SyncResponse) error {
	result, err := r.process(req)
	if err != nil {
		return err
	}

	item, ok := result.(*net.SyncResponse)
	if !ok {
		return ErrBadResult
	}
	*resp = *item
	return nil
}

// ForceSync handles force sync requests.
func (r *Lachesis) ForceSync(
	req *net.EagerSyncRequest, resp *net.EagerSyncResponse) error {
	result, err := r.process(req)
	if err != nil {
		return err
	}

	item, ok := result.(*net.EagerSyncResponse)
	if !ok {
		return ErrBadResult
	}
	*resp = *item
	return nil
}

// FastForward handles fast forward requests.
func (r *Lachesis) FastForward(
	req *net.FastForwardRequest, resp *net.FastForwardResponse) error {
	result, err := r.process(req)
	if err != nil {
		return err
	}

	item, ok := result.(*net.FastForwardResponse)
	if !ok {
		return ErrBadResult
	}
	*resp = *item
	return nil
}

func (r *Lachesis) send(req interface{}) *net.RPCResponse {
	reply := make(chan *net.RPCResponse, 1) // Buffered.
	ticket := &net.RPC{
		Command:  req,
		RespChan: reply,
	}

	timer := time.NewTimer(r.receiveTimeout)

	select {
	case r.receiver <- ticket:
	case <-timer.C:
		return &net.RPCResponse{Error: ErrReceiverIsBusy}
	case <-r.done:
		return &net.RPCResponse{Error: ErrTransportStopped}
	}

	var result *net.RPCResponse

	timer.Reset(r.processTimeout)

	select {
	case result = <-reply:
	case <-timer.C:
		result = &net.RPCResponse{Error: ErrProcessingTimeout}
	case <-r.done:
		return &net.RPCResponse{Error: ErrTransportStopped}
	}

	return result
}

func (r *Lachesis) process(req interface{}) (resp interface{}, err error) {
	result := r.send(req)
	if result.Error != nil {
		return nil, result.Error
	}
	resp = result.Response
	return
}
