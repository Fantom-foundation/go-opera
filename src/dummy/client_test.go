package dummy

import (
	"fmt"
	"testing"
	"time"

	bcrypto "github.com/andrecronje/lachesis/src/crypto"
	"github.com/stretchr/testify/assert"

	"github.com/andrecronje/lachesis/src/common"
	"github.com/andrecronje/lachesis/src/poset"
	"github.com/andrecronje/lachesis/src/proxy"
)

var (
	timeout    = 2 * time.Second
	errTimeout = "time is over"
)

func TestSocketProxyServer(t *testing.T) {
	assert := assert.New(t)
	addr := "127.0.0.1:9990"
	logger := common.NewTestLogger(t)

	tx_origin := []byte("the test transaction")

	// Server
	app, err := proxy.NewGrpcAppProxy(addr, timeout, logger)
	assert.NoError(err)

	//  listens for a request
	go func() {
		select {
		case tx := <-app.SubmitCh():
			assert.Equal(tx_origin, tx)
		case <-time.After(timeout):
			assert.Fail(errTimeout)
		}
	}()

	// Client part connecting to RPC service and calling methods
	lachesisProxy, err := proxy.NewGrpcLachesisProxy(addr, logger)
	assert.NoError(err)

	node, err := NewDummyClient(lachesisProxy, nil, logger)
	assert.NoError(err)

	err = node.SubmitTx(tx_origin)
	assert.NoError(err)
}

func TestDummySocketClient(t *testing.T) {
	assert := assert.New(t)
	addr := "127.0.0.1:9992"
	logger := common.NewTestLogger(t)

	// server
	appProxy, err := proxy.NewGrpcAppProxy(addr, timeout, logger)
	assert.NoError(err)
	defer appProxy.Close()

	// client
	lachesisProxy, err := proxy.NewGrpcLachesisProxy(addr, logger)
	assert.NoError(err)
	defer lachesisProxy.Close()

	state := NewState(logger)

	_, err = NewDummyClient(lachesisProxy, state, logger)
	assert.NoError(err)

	initialStateHash := state.stateHash
	//create a few blocks
	blocks := [5]poset.Block{}
	for i := 0; i < 5; i++ {
		blocks[i] = poset.NewBlock(i, i+1, []byte{}, [][]byte{[]byte(fmt.Sprintf("block %d transaction", i))})
	}

	<-time.After(timeout / 4)

	//commit first block and check that the client's statehash is correct
	stateHash, err := appProxy.CommitBlock(blocks[0])
	assert.NoError(err)

	expectedStateHash := initialStateHash

	for _, t := range blocks[0].Transactions() {
		tHash := bcrypto.SHA256(t)
		expectedStateHash = bcrypto.SimpleHashFromTwoHashes(expectedStateHash, tHash)
	}

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
