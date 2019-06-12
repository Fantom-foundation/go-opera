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
	"github.com/Fantom-foundation/go-lachesis/src/logger"
	"github.com/Fantom-foundation/go-lachesis/src/network"
	"github.com/Fantom-foundation/go-lachesis/src/poset"
	"github.com/Fantom-foundation/go-lachesis/src/proxy"
	"github.com/fortytw2/leaktest"
)

func TestSocketProxyServer(t *testing.T) {
	defer leaktest.CheckTimeout(t, time.Second)()

	const (
		timeout    = 2 * time.Second
		errTimeout = "time is over"
	)

	assertar := assert.New(t)
	lgr := common.NewTestLogger(t)

	// Server
	logger.SetTestMode(t)
	app, addr, err := proxy.NewGrpcAppProxy("server.fake", timeout, network.FakeListener)
	if !assertar.NoError(err) {
		return
	}
	defer app.Close()

	// Client part connecting to RPC service and calling methods
	dialer := network.FakeDialer("client.fake")
	lachesisProxy, err := proxy.NewGrpcLachesisProxy(addr, grpc.WithContextDialer(dialer))
	if !assertar.NoError(err) {
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
			assertar.Equal(txOrigin, tx)
		case <-time.After(timeout):
			assertar.Fail(errTimeout)
		}
	}()

	node, err := NewClient(lachesisProxy, nil, lgr)
	assertar.NoError(err)

	err = node.SubmitTx(txOrigin)
	assertar.NoError(err)

	wg.Wait()
}

func TestDummySocketClient(t *testing.T) {
	defer leaktest.CheckTimeout(t, time.Second)()

	const (
		timeout = 2 * time.Second
	)
	assertar := assert.New(t)
	lgr := common.NewTestLogger(t)

	// server
	logger.SetTestMode(t)
	appProxy, addr, err := proxy.NewGrpcAppProxy("server.fake", timeout, network.FakeListener)
	if !assertar.NoError(err) {
		return
	}
	defer appProxy.Close()

	// client
	dialer := network.FakeDialer("client.fake")
	lachesisProxy, err := proxy.NewGrpcLachesisProxy(addr, grpc.WithContextDialer(dialer))
	if !assertar.NoError(err) {
		return
	}
	defer lachesisProxy.Close()

	state := NewState(lgr)

	_, err = NewClient(lachesisProxy, state, lgr)
	assertar.NoError(err)

	initialStateHash := state.stateHash
	//create a few blocks
	blocks := [5]poset.Block{}
	for i := int64(0); i < 5; i++ {
		blocks[i] = poset.NewBlock(i, i+1, []byte{}, [][]byte{[]byte(fmt.Sprintf("block %d transaction", i))})
	}

	<-time.After(timeout / 4)

	//commit first block and check that the client's statehash is correct
	stateHash, err := appProxy.CommitBlock(blocks[0])
	assertar.NoError(err)

	expectedStateHash := crypto.Keccak256(append([][]byte{initialStateHash}, blocks[0].Transactions()...)...)
	assertar.Equal(expectedStateHash, stateHash)

	snapshot, err := appProxy.GetSnapshot(blocks[0].Index())
	assertar.NoError(err)
	assertar.Equal(expectedStateHash, snapshot)

	//commit a few more blocks, then attempt to restore back to block 0 state
	for i := 1; i < 5; i++ {
		_, err := appProxy.CommitBlock(blocks[i])
		assertar.NoError(err)
	}

	err = appProxy.Restore(snapshot)
	assertar.NoError(err)
}
