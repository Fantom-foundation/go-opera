package poset

import (
	"fmt"
	"testing"

	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	"github.com/Fantom-foundation/go-lachesis/src/peers"
	"github.com/Fantom-foundation/go-lachesis/src/pos"
	"github.com/Fantom-foundation/go-lachesis/src/state"
)

func TestStateBalances(t *testing.T) {
	// inmem store
	participants := peers.NewPeers()
	for i := 0; i < 3; i++ {
		key, _ := crypto.GenerateECDSAKey()
		pubKey := fmt.Sprintf("0x%X", crypto.FromECDSAPub(&key.PublicKey))
		p := peers.NewPeer(pubKey, fmt.Sprintf("addr%d", i))
		participants.AddPeer(p)
	}
	store := NewInmemStore(participants, 3, pos.DefaultConfig())

	roundStateDB := func(root common.Hash) *state.DB {
		db, err := state.New(root, store.StateDB())
		if err != nil {
			t.Fatal(err)
			return nil
		}
		return db
	}

	checkBalance := func(root common.Hash, addr common.Address, balance uint64) error {
		statedb := roundStateDB(root)
		got := statedb.GetBalance(addr)
		if got != balance {
			return fmt.Errorf("unexpected balance %d of %s at %s. %d expected", got, addr.String(), root.String(), balance)
		}
		return nil
	}

	var (
		err                error
		root, fork1, fork2 common.Hash

		aa = []common.Address{
			fakeAddress(0),
			fakeAddress(1),
			fakeAddress(2),
		}
	)

	// empty

	for _, a := range aa {
		err = checkBalance(root, a, 0)
		if err != nil {
			t.Fatal(err)
		}
	}

	// root

	statedb := roundStateDB(common.Hash{})
	statedb.AddBalance(aa[0], 10)
	root, err = statedb.Commit(true)
	if err != nil {
		t.Fatal(err)
	}

	// fork 1

	statedb = roundStateDB(root)
	statedb.AddBalance(aa[1], 11)
	fork1, err = statedb.Commit(true)
	if err != nil {
		t.Fatal(err)
	}
	if statedb.GetBalance(aa[1]) != 11 {
		t.Fatal("GetBalance() does not return actual before commit!")
	}

	// fork 2

	statedb = roundStateDB(root)
	statedb.AddBalance(aa[2], 12)
	statedb.SubBalance(aa[0], 5)
	fork2, err = statedb.Commit(true)
	if err != nil {
		t.Fatal(err)
	}

	// check root

	err = checkBalance(root, aa[0], 10)
	if err != nil {
		t.Fatal(err)
	}

	err = checkBalance(root, aa[1], 0)
	if err != nil {
		t.Fatal(err)
	}

	err = checkBalance(root, aa[2], 0)
	if err != nil {
		t.Fatal(err)
	}

	// check fork1

	err = checkBalance(fork1, aa[0], 10)
	if err != nil {
		t.Fatal(err)
	}

	err = checkBalance(fork1, aa[1], 11)
	if err != nil {
		t.Fatal(err)
	}

	err = checkBalance(fork1, aa[2], 0)
	if err != nil {
		t.Fatal(err)
	}

	// check fork2

	err = checkBalance(fork2, aa[0], 5)
	if err != nil {
		t.Fatal(err)
	}

	err = checkBalance(fork2, aa[1], 0)
	if err != nil {
		t.Fatal(err)
	}

	err = checkBalance(fork2, aa[2], 12)
	if err != nil {
		t.Fatal(err)
	}

}

/*
 * Staff:
 */

func fakeAddress(n int64) (h common.Address) {
	for i := 8; i >= 1; i-- {
		h[i-1] = byte(n)
		n = n >> 8
	}
	return
}
