package net

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/andrecronje/lachesis/src/poset"
)

func TestInmemTransport(t *testing.T) {
	timeout := 200 * time.Millisecond

	// Transport 1 is consumer
	_, trans1 := NewInmemTransport("")
	defer trans1.Close()

	rpcCh := trans1.Consumer()

	// Transport 2 makes outbound request
	_, trans2 := NewInmemTransport("")
	defer trans2.Close()

	t.Run("Sync", func(t *testing.T) {
		assert := assert.New(t)

		expectedReq := &SyncRequest{
			FromID: 0,
			Known: map[int64]int64{
				0: 1,
				1: 2,
				2: 3,
			},
		}

		expectedResp := &SyncResponse{
			FromID: 1,
			Events: []poset.WireEvent{
				poset.WireEvent{
					Body: poset.WireBody{
						Transactions:         [][]byte(nil),
						SelfParentIndex:      1,
						OtherParentCreatorID: 10,
						OtherParentIndex:     0,
						CreatorID:            9,
					},
				},
			},
			Known: map[int64]int64{
				0: 5,
				1: 5,
				2: 6,
			},
		}

		go func() {
			select {
			case rpc := <-rpcCh:
				req := rpc.Command.(*SyncRequest)
				assert.EqualValues(expectedReq, req)
				rpc.Respond(expectedResp, nil)
			case <-time.After(timeout):
				assert.Fail("timeout")
			}
		}()

		var resp = new(SyncResponse)
		err := trans2.Sync(trans1.LocalAddr(), expectedReq, resp)
		if assert.NoError(err) {
			assert.EqualValues(expectedResp, resp)
		}
	})

	t.Run("EagerSync", func(t *testing.T) {
		assert := assert.New(t)

		expectedReq := &EagerSyncRequest{
			FromID: 0,
			Events: []poset.WireEvent{
				poset.WireEvent{
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

		expectedResp := &EagerSyncResponse{
			FromID:  1,
			Success: true,
		}

		go func() {
			select {
			case rpc := <-rpcCh:
				req := rpc.Command.(*EagerSyncRequest)
				assert.EqualValues(expectedReq, req)
				rpc.Respond(expectedResp, nil)
			case <-time.After(timeout):
				assert.Fail("timeout")
			}
		}()

		var resp = new(EagerSyncResponse)
		err := trans2.EagerSync(trans1.LocalAddr(), expectedReq, resp)
		if assert.NoError(err) {
			assert.EqualValues(expectedResp, resp)
		}
	})

	t.Run("FastForward", func(t *testing.T) {
		assert := assert.New(t)

		expectedReq := &FastForwardRequest{
			FromID: 0,
		}

		frame := poset.Frame{}
		block, err := poset.NewBlockFromFrame(1, frame)
		assert.NoError(err)
		expectedResp := &FastForwardResponse{
			FromID:   1,
			Block:    block,
			Frame:    frame,
			Snapshot: []byte("snapshot"),
		}

		go func() {
			select {
			case rpc := <-rpcCh:
				req := rpc.Command.(*FastForwardRequest)
				assert.EqualValues(expectedReq, req)
				rpc.Respond(expectedResp, nil)
			case <-time.After(timeout):
				assert.Fail("timeout")
			}
		}()

		var resp = new(FastForwardResponse)
		err = trans2.FastForward(trans1.LocalAddr(), expectedReq, resp)
		if assert.NoError(err) {
			assert.EqualValues(expectedResp, resp)
		}
	})
}
