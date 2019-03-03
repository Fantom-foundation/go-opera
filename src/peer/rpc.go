package peer

import (
	"time"
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
	receiver       chan *RPC
	processTimeout time.Duration
	receiveTimeout time.Duration
}

// NewLachesis creates new Lachesis RPC handler.
func NewLachesis(done chan struct{}, receiver chan *RPC,
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
	req *SyncRequest, resp *SyncResponse) error {
	result, err := r.process(req)
	if err != nil {
		return err
	}

	item, ok := result.(*SyncResponse)
	if !ok {
		return ErrBadResult
	}
	*resp = *item
	return nil
}

// ForceSync handles force sync requests.
func (r *Lachesis) ForceSync(
	req *ForceSyncRequest, resp *ForceSyncResponse) error {
	result, err := r.process(req)
	if err != nil {
		return err
	}

	item, ok := result.(*ForceSyncResponse)
	if !ok {
		return ErrBadResult
	}
	*resp = *item
	return nil
}

// FastForward handles fast forward requests.
func (r *Lachesis) FastForward(
	req *FastForwardRequest, resp *FastForwardResponse) error {
	result, err := r.process(req)
	if err != nil {
		return err
	}

	item, ok := result.(*FastForwardResponse)
	if !ok {
		return ErrBadResult
	}
	*resp = *item
	return nil
}

func (r *Lachesis) send(req interface{}) *RPCResponse {
	reply := make(chan *RPCResponse, 1) // Buffered.
	ticket := &RPC{
		Command:  req,
		RespChan: reply,
	}

	timer := time.NewTimer(r.receiveTimeout)

	select {
	case r.receiver <- ticket:
	case <-timer.C:
		return &RPCResponse{Error: ErrReceiverIsBusy}
	case <-r.done:
		return &RPCResponse{Error: ErrTransportStopped}
	}

	var result *RPCResponse

	timer.Reset(r.processTimeout)

	select {
	case result = <-reply:
	case <-timer.C:
		result = &RPCResponse{Error: ErrProcessingTimeout}
	case <-r.done:
		return &RPCResponse{Error: ErrTransportStopped}
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
