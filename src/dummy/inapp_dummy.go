package dummy

import (
	"time"

  "github.com/andrecronje/lachesis/src/dummy/state"
  "github.com/andrecronje/lachesis/src/proxy/inapp"
	"github.com/sirupsen/logrus"
)

// DummyInappClient is an in-memory implmentation of the dummy app. It actually
// imlplements the AppProxy interface, and can be passed in the Lachesis
// constructor directly
type DummyInappClient struct {
	*inapp.InappProxy
	state  *state.State
	logger *logrus.Logger
}

//NewDummyInappClient instantiates a DummyInappClient
func NewDummyInappClient(logger *logrus.Logger) *DummyInappClient {
	state := state.NewState(logger)
	proxy := inapp.NewInappProxy(1*time.Second, logger)
  client := &DummyInappClient{
		InappProxy: proxy,
		state:      state,
		logger:     logger,
	}
 	return client
}

//SubmitTx sends a transaction to the Lachesis node via the InappProxy
func (c *DummyInappClient) SubmitTx(tx []byte) error {
	return c.InappProxy.SubmitTx(tx)
}

//GetCommittedTransactions returns the state's list of transactions
func (c *DummyInappClient) GetCommittedTransactions() [][]byte {
	return c.state.GetCommittedTransactions()
}
