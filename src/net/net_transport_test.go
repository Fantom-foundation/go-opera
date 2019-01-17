package net

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/poset"
)

func TestNetworkTransport(t *testing.T) {
	logger := common.NewTestLogger(t)
	maxPool := 3

	// Transport 1 is consumer
	trans1, err := NewTCPTransport("127.0.0.1:0", nil, 2, time.Second, logger)
	assert.NoError(t, err)
	defer trans1.Close()

	// Transport 2 makes outbound request
	trans2, err := NewTCPTransport("127.0.0.1:0", nil, maxPool, time.Second, logger)
	assert.NoError(t, err)
	defer trans2.Close()

	testTransportImplementation(t, trans1, trans2)

	rpcCh := trans1.Consumer()

	t.Run("PooledConn", func(t *testing.T) {
		assert := assert.New(t)

		expectedReq := &SyncRequest{
			FromID: 0,
			Known: map[uint64]int64{
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
			Known: map[uint64]int64{
				0: 5,
				1: 5,
				2: 6,
			},
		}

		go func() {
			for {
				select {
				case rpc := <-rpcCh:
					req := rpc.Command.(*SyncRequest)
					assert.EqualValues(expectedReq, req)
					rpc.Respond(expectedResp, nil)
				case <-time.After(200 * time.Millisecond):
					return
				}
			}
		}()

		wg := &sync.WaitGroup{}

		appendFunc := func() {
			defer wg.Done()
			var resp = new(SyncResponse)
			err := trans2.Sync(trans1.LocalAddr(), expectedReq, resp)
			if assert.NoError(err) {
				assert.EqualValues(expectedResp, resp)
			}
		}

		// Try to do parallel appends, should stress the conn pool
		count := maxPool * 2
		wg.Add(count)
		for i := 0; i < count; i++ {
			go appendFunc()
		}
		wg.Wait()

		// Check the conn pool size
		addr := trans1.LocalAddr()
		assert.Equal(maxPool, len(trans2.connPool[addr]))
	})
}
