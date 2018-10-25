package socket

import (
	"reflect"
	"testing"
	"time"

	"github.com/andrecronje/lachesis/src/common"
	"github.com/andrecronje/lachesis/src/poset"
	aproxy "github.com/andrecronje/lachesis/src/proxy/socket/app"
	bproxy "github.com/andrecronje/lachesis/src/proxy/socket/lachesis"
	"github.com/sirupsen/logrus"
)

type TestHandler struct {
	blocks     []poset.Block
	blockIndex int
	snapshot   []byte
	logger     *logrus.Logger
}

func (p *TestHandler) CommitHandler(block poset.Block) ([]byte, error) {
	p.logger.Debug("CommitBlock")

	p.blocks = append(p.blocks, block)

	return []byte("statehash"), nil
}

func (p *TestHandler) SnapshotHandler(blockIndex int) ([]byte, error) {
	p.logger.Debug("GetSnapshot")

	p.blockIndex = blockIndex

	return []byte("snapshot"), nil
}

func (p *TestHandler) RestoreHandler(snapshot []byte) ([]byte, error) {
	p.logger.Debug("RestoreSnapshot")

	p.snapshot = snapshot

	return []byte("statehash"), nil
}

func NewTestHandler(t *testing.T) *TestHandler {
	logger := common.NewTestLogger(t)

	return &TestHandler{
		blocks:     []poset.Block{},
		blockIndex: 0,
		snapshot:   []byte{},
		logger:     logger,
	}
}

func TestSocketProxyServer(t *testing.T) {
	//clientAddr := "127.0.0.1:9990"
	proxyAddr := "127.0.0.1:9991"

	appProxy, err := aproxy.NewWebsocketAppProxy(proxyAddr, 1*time.Second, common.NewTestLogger(t))

	if err != nil {
		t.Fatalf("Cannot create SocketAppProxy: %s", err)
	}

	submitCh := appProxy.SubmitCh()

	time.Sleep(time.Millisecond * 5) // give chance for ws conn to establish

	tx := []byte("the test transaction")

	// Listen for a request
	go func() {
		select {
		case st := <-submitCh:
			// Verify the command
			if !reflect.DeepEqual(st, tx) {
				t.Fatalf("tx mismatch: %#v %#v", tx, st)
			}
		case <-time.After(200 * time.Millisecond):
			t.Fatalf("timeout")
		}
	}()

	// now client part connecting to RPC service
	// and calling methods
	lachesisProxy, err := bproxy.NewWebsocketLachesisProxy(proxyAddr, NewTestHandler(t), 1*time.Second, common.NewTestLogger(t))

	if err != nil {
		t.Fatal(err)
	}

	err = lachesisProxy.SubmitTx(tx)

	if err != nil {
		t.Fatal(err)
	}
}

func TestSocketProxyClient(t *testing.T) {
	//clientAddr := "127.0.0.1:9992"
	proxyAddr := "127.0.0.1:9993"

	logger := common.NewTestLogger(t)

	//create app proxy
	appProxy, err := aproxy.NewWebsocketAppProxy(proxyAddr, 1*time.Second, logger)
	if err != nil {
		t.Fatalf("Cannot create SocketAppProxy: %s", err)
	}

	handler := NewTestHandler(t)

	//create lachesis proxy
	_, err = bproxy.NewWebsocketLachesisProxy(proxyAddr, handler, 1*time.Second, logger)

	transactions := [][]byte{
		[]byte("tx 1"),
		[]byte("tx 2"),
		[]byte("tx 3"),
	}

	block := poset.NewBlock(0, 1, []byte{}, transactions)
	expectedStateHash := []byte("statehash")
	expectedSnapshot := []byte("snapshot")

	// TODO: Drain CommitCh on lachesis proxy
	stateHash, err := appProxy.CommitBlock(block)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(block, handler.blocks[0]) {
		t.Fatalf("block should be %v, not %v", block, handler.blocks[0])
	}

	if !reflect.DeepEqual(stateHash, expectedStateHash) {
		t.Fatalf("StateHash should be %v, not %v", expectedStateHash, stateHash)
	}

	snapshot, err := appProxy.GetSnapshot(block.Index())
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(block.Index(), handler.blockIndex) {
		t.Fatalf("blockIndex should be %v, not %v", block.Index(), handler.blockIndex)
	}

	if !reflect.DeepEqual(snapshot, expectedSnapshot) {
		t.Fatalf("Snapshot should be %v, not %v", expectedSnapshot, snapshot)
	}

	err = appProxy.Restore(snapshot)
	if err != nil {
		t.Fatalf("Error restoring snapshot: %v", err)
	}

	if !reflect.DeepEqual(expectedSnapshot, handler.snapshot) {
		t.Fatalf("snapshot should be %v, not %v", expectedSnapshot, handler.snapshot)
	}

}
