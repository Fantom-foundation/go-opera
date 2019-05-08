package dummy

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	"github.com/Fantom-foundation/go-lachesis/src/network"
	"github.com/Fantom-foundation/go-lachesis/src/poset"
	"github.com/Fantom-foundation/go-lachesis/src/proxy"
)

func TestSocketProxyServer(t *testing.T) {
	const (
		timeout    = 2 * time.Second
		errTimeout = "time is over"
	)

	assert := assert.New(t)
	logger := common.NewTestLogger(t)

	// Server
	app, addr, err := proxy.NewGrpcAppProxy("server.fake", timeout, logger, network.FakeListener)
	if !assert.NoError(err) {
		return
	}
	defer app.Close()

	// Client part connecting to RPC service and calling methods
	dialer := network.FakeDialer("client.fake")
	lachesisProxy, err := proxy.NewGrpcLachesisProxy(addr, logger, grpc.WithContextDialer(dialer))
	if !assert.NoError(err) {
		return
	}
	defer lachesisProxy.Close()

	txOrigin := []byte("the test transaction")

	//  listens for a request
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		select {
		case tx := <-app.SubmitCh():
			assert.Equal(txOrigin, tx)
		case <-time.After(timeout):
			assert.Fail(errTimeout)
		}
	}()

	node, err := NewDummyClient(lachesisProxy, nil, logger)
	assert.NoError(err)

	err = node.SubmitTx(txOrigin)
	assert.NoError(err)

	wg.Wait()
}

func TestDummySocketClient(t *testing.T) {
	const (
		timeout = 2 * time.Second
	)
	assert := assert.New(t)
	logger := common.NewTestLogger(t)

	// server
	appProxy, addr, err := proxy.NewGrpcAppProxy("server.fake", timeout, logger, network.FakeListener)
	if !assert.NoError(err) {
		return
	}
	defer appProxy.Close()

	// client
	dialer := network.FakeDialer("client.fake")
	lachesisProxy, err := proxy.NewGrpcLachesisProxy(addr, logger, grpc.WithContextDialer(dialer))
	if !assert.NoError(err) {
		return
	}
	defer lachesisProxy.Close()

	state := NewState(logger)

	_, err = NewDummyClient(lachesisProxy, state, logger)
	assert.NoError(err)

	initialStateHash := state.stateHash
	//create a few blocks
	blocks := [5]poset.Block{}
	for i := int64(0); i < 5; i++ {
		blocks[i] = poset.NewBlock(i, i+1, []byte{}, [][]byte{[]byte(fmt.Sprintf("block %d transaction", i))})
	}

	<-time.After(timeout / 4)

	//commit first block and check that the client's statehash is correct
	stateHash, err := appProxy.CommitBlock(blocks[0])
	assert.NoError(err)

	expectedStateHash := crypto.Keccak256(append([][]byte{initialStateHash}, blocks[0].Transactions()...)...)
	assert.Equal(expectedStateHash, stateHash)

	snapshot, err := appProxy.GetSnapshot(blocks[0].Index())
	assert.NoError(err)
	assert.Equal(expectedStateHash, snapshot)

	//commit a few more blocks, then attempt to restore back to block 0 state
	for i := 1; i < 5; i++ {
		_, err := appProxy.CommitBlock(blocks[i])
		assert.NoError(err)
	}

	err = appProxy.Restore(snapshot)
	assert.NoError(err)
}
