package dummy

import (
	"github.com/andrecronje/lachesis/src/poset"
	"github.com/andrecronje/lachesis/src/proxy/inmem"
	"github.com/andrecronje/lachesis/src/dummy/state"
	"github.com/sirupsen/logrus"
)

// InmemDummyClient is an in-memory implementation of the dummy app. It actually
// imlplements the AppProxy interface, and can be passed in the Babble
// constructor directly
type InmemDummyClient struct {
	*inmem.InmemProxy
	state  *state.State
	logger *logrus.Logger
}

//NewInmemDummyClient instantiates an InemDummyClient
func NewInmemDummyClient(logger *logrus.Logger) *InmemDummyClient {
 	state := state.NewState(logger)
 	commitHandler := func(block poset.Block) ([]byte, error) {
		logger.Debug("CommitBlock")
		return state.CommitBlock(block)
	}
 	snapshotHandler := func(blockIndex int) ([]byte, error) {
		logger.Debug("GetSnapshot")
		return state.GetSnapshot(blockIndex)
	}
 	restoreHandler := func(snapshot []byte) ([]byte, error) {
		logger.Debug("RestoreSnapshot")
		return state.Restore(snapshot)
	}
 	proxy := inmem.NewInmemProxy(commitHandler, snapshotHandler, restoreHandler, logger)
 	client := &InmemDummyClient{
		InmemProxy: proxy,
		state:      state,
		logger:     logger,
	}
 	return client
}

//SubmitTx sends a transaction to the Lachesis node via the InappProxy
func (c *InmemDummyClient) SubmitTx(tx []byte) {
	c.InmemProxy.SubmitTx(tx)
}

//GetCommittedTransactions returns the state's list of transactions
func (c *InmemDummyClient) GetCommittedTransactions() [][]byte {
	return c.state.GetCommittedTransactions()
}
