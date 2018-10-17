package inmem

import (
	"reflect"
	"testing"
	"time"
 	"github.com/andrecronje/lachesis/src/common"
	"github.com/andrecronje/lachesis/src/poset"
)

type TestProxy struct {
	*InmemProxy
	transactions [][]byte
}

func NewTestProxy(t *testing.T) *TestProxy {
	logger := common.NewTestLogger(t)
 	proxy := &TestProxy{
		transactions: [][]byte{},
	}
 	commitHandler := func(block poset.Block) ([]byte, error) {
		logger.Debug("CommitBlock")
		proxy.transactions = append(proxy.transactions, block.Transactions()...)
		return []byte("statehash"), nil
	}
 	snapshotHandler := func(blockIndex int) ([]byte, error) {
		logger.Debug("GetSnapshot")
		return []byte("snapshot"), nil
	}
 	restoreHandler := func(snapshot []byte) ([]byte, error) {
		logger.Debug("RestoreSnapshot")
		return []byte("statehash"), nil
	}
 	proxy.InmemProxy = NewInmemProxy(commitHandler, snapshotHandler, restoreHandler, logger)
 	return proxy
}

func TestInmemProxyAppSide(t *testing.T) {
	proxy := NewTestProxy(t)
 	submitCh := proxy.SubmitCh()
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
 	proxy.SubmitTx(tx)
}

func TestInmemProxyLachesisSide(t *testing.T) {
	proxy := NewTestProxy(t)
 	transactions := [][]byte{
		[]byte("tx 1"),
		[]byte("tx 2"),
		[]byte("tx 3"),
	}
 	block := poset.NewBlock(0, 1, []byte{}, transactions)
 	/***************************************************************************
	Commit
	***************************************************************************/
	stateHash, err := proxy.CommitBlock(block)
	if err != nil {
		t.Fatal(err)
	}
 	expectedStateHash := []byte("statehash")
	if !reflect.DeepEqual(stateHash, expectedStateHash) {
		t.Fatalf("StateHash should be %v, not %v", expectedStateHash, stateHash)
	}
 	if !reflect.DeepEqual(transactions, proxy.transactions) {
		t.Fatalf("Transactions should be %v, not %v", transactions, proxy.transactions)
	}
 	/***************************************************************************
	Snapshot
	***************************************************************************/
	snapshot, err := proxy.GetSnapshot(block.Index())
	if err != nil {
		t.Fatal(err)
	}
 	expectedSnapshot := []byte("snapshot")
	if !reflect.DeepEqual(snapshot, expectedSnapshot) {
		t.Fatalf("Snapshot should be %v, not %v", expectedSnapshot, snapshot)
	}
 	/***************************************************************************
	Restore
	***************************************************************************/
 	stateHash, err = proxy.Restore(snapshot)
	if err != nil {
		t.Fatalf("Error restoring snapshot: %v", err)
	}
 	if !reflect.DeepEqual(stateHash, expectedStateHash) {
		t.Fatalf("StateHash should be %v, not %v", expectedStateHash, stateHash)
	}
}
