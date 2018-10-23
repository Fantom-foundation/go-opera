package dummy

import (
	"fmt"
	"reflect"
	"testing"
	"time"
 	"github.com/andrecronje/lachesis/src/common"
	bcrypto "github.com/andrecronje/lachesis/src/crypto"
	"github.com/andrecronje/lachesis/src/poset"
)

func TestInappDummySubmit(t *testing.T) {
 	dummy := NewDummyInappClient(common.NewTestLogger(t))
 	submitCh := dummy.SubmitCh()
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
	err := dummy.SubmitTx(tx)
 	if err != nil {
		t.Fatal(err)
	}
}

func TestInappDummyCommitAndSnapshots(t *testing.T) {
 	dummy := NewDummyInappClient(common.NewTestLogger(t))
 	initialStateHash := []byte{}
 	//create a few blocks
	blocks := [5]poset.Block{}
	for i := 0; i < 5; i++ {
		blocks[i] = poset.NewBlock(i, i+1, []byte{}, [][]byte{[]byte(fmt.Sprintf("block %d transaction", i))})
	}
 	//commit first block and check that the client's statehash is correct
	stateHash, err := dummy.proxy.CommitBlock(blocks[0])
	if err != nil {
		t.Fatal(err)
	}
 	expectedStateHash := initialStateHash
 	for _, t := range blocks[0].Transactions() {
		tHash := bcrypto.SHA256(t)
 		expectedStateHash = bcrypto.SimpleHashFromTwoHashes(expectedStateHash, tHash)
	}
 	if !reflect.DeepEqual(stateHash, expectedStateHash) {
		t.Fatalf("StateHash should be %v, not %v", expectedStateHash, stateHash)
	}
 	snapshot, err := dummy.proxy.GetSnapshot(blocks[0].Index())
 	if err != nil {
		t.Fatal(err)
	}
 	if !reflect.DeepEqual(snapshot, expectedStateHash) {
		t.Fatalf("Snapshot should be %v, not %v", expectedStateHash, snapshot)
	}
 	//commit a few more blocks, then attempt to restore back to block 0 state
	for i := 1; i < 5; i++ {
		_, err := dummy.proxy.CommitBlock(blocks[i])
 		if err != nil {
			t.Fatal(err)
		}
	}
 	err = dummy.proxy.Restore(snapshot)
 	if err != nil {
		t.Fatalf("Error restoring snapshot: %v", err)
	}
 }
