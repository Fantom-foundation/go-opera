package proxy

import (
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/andrecronje/lachesis/src/common"
	"github.com/andrecronje/lachesis/src/poset"
)

var (
	timeout    = 200 * time.Millisecond
	errTimeout = "time is over"
)

func TestInmemAppCalls(t *testing.T) {
	proxy := NewTestProxy(t)

	transactions := [][]byte{
		[]byte("tx 1"),
		[]byte("tx 2"),
		[]byte("tx 3"),
	}
	block := poset.NewBlock(0, 1, []byte{}, transactions)

	t.Run("#1 Send tx", func(t *testing.T) {
		assert := assert.New(t)

		tx_origin := []byte("the test transaction")

		go func() {
			select {
			case tx := <-proxy.SubmitCh():
				assert.Equal(tx_origin, tx)
			case <-time.After(timeout):
				assert.Fail(errTimeout)
			}
		}()

		proxy.SubmitTx(tx_origin)
	})

	t.Run("#2 Commit block", func(t *testing.T) {
		assert := assert.New(t)

		stateHash, err := proxy.CommitBlock(block)
		assert.NoError(err)
		assert.EqualValues(goldStateHash(), stateHash)
		assert.EqualValues(transactions, proxy.transactions)
	})

	t.Run("#3 Get snapshot", func(t *testing.T) {
		assert := assert.New(t)

		snapshot, err := proxy.GetSnapshot(block.Index())
		assert.NoError(err)
		assert.Equal(goldSnapshot(), snapshot)
	})

	t.Run("#4 Restore snapshot", func(t *testing.T) {
		assert := assert.New(t)

		err := proxy.Restore(goldSnapshot())
		assert.NoError(err)
	})
}

/*
 * staff
 */

type TestProxy struct {
	*InmemAppProxy
	transactions [][]byte
	logger       *logrus.Logger
}

func NewTestProxy(t *testing.T) *TestProxy {
	proxy := &TestProxy{
		transactions: [][]byte{},
		logger:       common.NewTestLogger(t),
	}

	proxy.InmemAppProxy = NewInmemAppProxy(proxy, proxy.logger)

	return proxy
}

func (p *TestProxy) CommitHandler(block poset.Block) ([]byte, error) {
	p.logger.Debug("CommitBlock")
	p.transactions = append(p.transactions, block.Transactions()...)
	return goldStateHash(), nil
}

func (p *TestProxy) SnapshotHandler(blockIndex int) ([]byte, error) {
	p.logger.Debug("GetSnapshot")
	return goldSnapshot(), nil
}

func (p *TestProxy) RestoreHandler(snapshot []byte) ([]byte, error) {
	p.logger.Debug("RestoreSnapshot")
	return goldStateHash(), nil
}

func goldStateHash() []byte {
	return []byte("statehash")
}

func goldSnapshot() []byte {
	return []byte("snapshot")
}
