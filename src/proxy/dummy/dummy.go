package dummy
 import (
	"time"
 	bproxy "github.com/andrecronje/lachesis/src/proxy/socket/lachesis"
	"github.com/sirupsen/logrus"
)
 type DummySocketClient struct {
	state         *State
	lachesisProxy *bproxy.SocketLachesisProxy
	logger        *logrus.Logger
}
 func NewDummySocketClient(clientAddr string, nodeAddr string, logger *logrus.Logger) (*DummySocketClient, error) {
 	lachesisProxy, err := bproxy.NewSocketLachesisProxy(nodeAddr, clientAddr, 1*time.Second, logger)
	if err != nil {
		return nil, err
	}
 	state := State{
		stateHash: []byte{},
		snapshots: make(map[int][]byte),
		logger:    logger,
	}
	state.writeMessage([]byte(clientAddr))
 	client := &DummySocketClient{
		state:       &state,
		lachesisProxy: lachesisProxy,
		logger:      logger,
	}
 	go client.Run()
 	return client, nil
}
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
 func (c *DummySocketClient) SubmitTx(tx []byte) error {
	return c.lachesisProxy.SubmitTx(tx)
}
