package dummy

import (
	"time"
  "github.com/andrecronje/lachesis/src/dummy/state"
	socket "github.com/andrecronje/lachesis/src/proxy/socket/lachesis"
	"github.com/sirupsen/logrus"
)

// DummySocketClient is a socket implementation of the dummy app. Lachesis and the
// app run in separate processes and communicate through TCP sockets using
// a SocketLachesisProxy and a SocketAppProxy.
type DummySocketClient struct {
	state         *state.State
	lachesisProxy *socket.SocketLachesisProxy
	logger        *logrus.Logger
}

// NewDummySocketClient instantiates a DummySocketClient and starts the
// SocketLachesisProxy
func NewDummySocketClient(clientAddr string, nodeAddr string, logger *logrus.Logger) (*DummySocketClient, error) {
 	lachesisProxy, err := socket.NewSocketLachesisProxy(nodeAddr, clientAddr, 1*time.Second, logger)
	if err != nil {
		return nil, err
	}
 	state := state.NewState(logger)
 	client := &DummySocketClient{
		state:         state,
		lachesisProxy: lachesisProxy,
		logger:        logger,
	}
 	go client.Run()
 	return client, nil
}

//Run listens for messages from Lachesis via the SocketProxy
func (c *DummySocketClient) Run() {
	for {
		select {
		case commit := <-c.lachesisProxy.CommitCh():
			c.logger.Debug("CommitBlock")
			stateHash, err := c.state.CommitBlock(commit.Block)
			commit.Respond(stateHash, err)
		case snapshotRequest := <-c.lachesisProxy.SnapshotRequestCh():
			c.logger.Debug("GetSnapshot")
			snapshot, err := c.state.GetSnapshot(snapshotRequest.BlockIndex)
			snapshotRequest.Respond(snapshot, err)
		case restoreRequest := <-c.lachesisProxy.RestoreCh():
			c.logger.Debug("Restore")
			stateHash, err := c.state.Restore(restoreRequest.Snapshot)
			restoreRequest.Respond(stateHash, err)
		}
	}
}

//SubmitTx sends a transaction to Babble via the SocketProxy
func (c *DummySocketClient) SubmitTx(tx []byte) error {
	return c.lachesisProxy.SubmitTx(tx)
}
