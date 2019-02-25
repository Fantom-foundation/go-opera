package peer

import (
	"context"
	"github.com/Fantom-foundation/go-lachesis/src/poset"
	"github.com/sirupsen/logrus"
)

// SyncRequest initiates a synchronization request.
type SyncRequest struct {
	FromID uint64
	Known  map[uint64]int64
}

// SyncResponse is a response to a SyncRequest request.
type SyncResponse struct {
	FromID    uint64
	SyncLimit bool
	Events    []poset.WireEvent
	Known     map[uint64]int64
}

// ForceSyncRequest after an initial sync to quickly catch up.
type ForceSyncRequest struct {
	FromID uint64
	Events []poset.WireEvent
}

// ForceSyncResponse response to an ForceSyncRequest.
type ForceSyncResponse struct {
	FromID  uint64
	Success bool
}

// FastForwardRequest request to start a fast forward catch up.
type FastForwardRequest struct {
	FromID uint64
}

// FastForwardResponse response with the snapshot data for fast forward
// request.
type FastForwardResponse struct {
	FromID   uint64
	Block    poset.Block
	Frame    poset.Frame
	Snapshot []byte
}

// RPCResponse captures both a response and a potential error.
type RPCResponse struct {
	Response interface{}
	Error    error
}

// RPC has a command, and provides a response mechanism.
type RPC struct {
	Command  interface{}
	RespChan chan<- *RPCResponse
}

// SendResult sends a result of a request.
func (rpc *RPC) SendResult(ctx context.Context,
	logger logrus.FieldLogger, result interface{}, err error) {
	logger = logger.WithField("method", "SendResult")
	select {
	case rpc.RespChan <- &RPCResponse{Response: result, Error: err}:
	case <-ctx.Done():
		logger.Warn(ctx.Err())
	}
}
