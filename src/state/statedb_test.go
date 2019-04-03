package state

import (
	"fmt"
	"testing"

	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/kvdb"
)

func TestStateDB(t *testing.T) {
	mem := kvdb.NewMemDatabase()
	store := NewDatabase(mem)

	stateAt := func(point common.Hash) *DB {
		db, err := New(point, store)
		if err != nil {
			t.Fatal(err)
			return nil
		}
		return db
	}

	checkBalance := func(point common.Hash, addr common.Address, balance uint64) error {
		db := stateAt(point)
		got := db.GetBalance(addr)
		if got != balance {
			return fmt.Errorf("unexpected balance %d of %s at %s. %d expected", got, addr.String(), point.String(), balance)
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

	db := stateAt(common.Hash{})
	db.AddBalance(aa[0], 10)
	root, err = db.Commit(true)
	if err != nil {
		t.Fatal(err)
	}

	// fork 1

	db = stateAt(root)
	db.AddBalance(aa[1], 11)
	fork1, err = db.Commit(true)
	if err != nil {
		t.Fatal(err)
	}
	if db.GetBalance(aa[1]) != 11 {
		t.Fatal("GetBalance() does not return actual before commit!")
	}

	// fork 2

	db = stateAt(root)
	db.AddBalance(aa[2], 12)
	db.SubBalance(aa[0], 5)
	fork2, err = db.Commit(true)
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
