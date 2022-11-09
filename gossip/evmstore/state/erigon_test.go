package state

import (
	"bytes"
	"math/big"
	"math/rand"
	"testing"

	"github.com/ledgerwatch/erigon-lib/kv/memdb"
	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	ecommon "github.com/ledgerwatch/erigon/common"
	estate "github.com/ledgerwatch/erigon/core/state"
)

func TestSoftFinalize(t *testing.T) {

	_, tx := memdb.NewTestTx(t)
	defer tx.Rollback()

	r := estate.NewPlainStateReader(tx)
	w := estate.NewPlainStateWriterNoHistory(tx)
	state := NewWithStateReader(r)

	_, addr1 := randomStateAccount(t, false)

	require.Len(t, state.journal.dirties, 0)
	require.Len(t, state.stateObjects, 0)
	state.getStateObject(addr1)
	require.Len(t, state.stateObjects, 1)

	require.Len(t, state.journal.dirties, 1)
	require.Len(t, state.stateObjectsDirty, 0)
	require.NoError(t, state.SoftFinalize(w))
	require.Len(t, state.journal.dirties, 0)
	require.Len(t, state.stateObjectsDirty, 1)
	delete(state.stateObjects, addr1)
	delete(state.stateObjectsDirty, addr1)

	_, addr2 := randomStateAccount(t, false)
	state.SetNonce(addr2, 20)
	require.Len(t, state.journal.dirties, 1)
	require.Len(t, state.stateObjects, 1)
	delete(state.stateObjects, addr2)
	require.Len(t, state.stateObjects, 0)
	require.Len(t, state.stateObjectsDirty, 0)
	require.NoError(t, state.SoftFinalize(w))
	require.Len(t, state.journal.dirties, 0)
	require.Len(t, state.stateObjectsDirty, 0)
}

func TestUpdateAccount(t *testing.T) {
	_, tx := memdb.NewTestTx(t)
	defer tx.Rollback()

	r := estate.NewPlainStateReader(tx)
	w := estate.NewPlainStateWriterNoHistory(tx)
	state := NewWithStateReader(r)

	_, addr := randomStateAccount(t, false)

	// dirty and empty, call DeleteAccount
	so := state.getStateObject(addr)
	require.Equal(t, so.Nonce(), uint64(0))
	require.Equal(t, so.data.Balance.Sign(), 0)
	require.True(t, bytes.Equal(so.data.CodeHash, emptyCodeHash))
	require.True(t, so.empty())
	require.NoError(t, updateAccount(w, addr, so, true))
	require.True(t, so.deleted)
	acc, err := state.stateReader.ReadAccountData(ecommon.Address(addr))
	require.Nil(t, acc)
	require.NoError(t, err)

	// dirty and no empty
	_, addr1 := randomStateAccount(t, false)
	so1 := state.getStateObject(addr1)
	state.SetNonce(addr1, 20)
	require.False(t, so1.empty())
	require.NotNil(t, so1)
	require.NoError(t, updateAccount(w, addr, so1, true))
	require.False(t, so1.deleted)
	acc1, err := state.stateReader.ReadAccountData(ecommon.Address(addr1))
	require.Nil(t, acc1)
	require.NoError(t, err)

	// isDirty, not empty and not suicided , created false, test UpdateAccountData
	_, addr2 := randomStateAccount(t, false)
	so2 := state.getStateObject(addr2)
	require.NoError(t, updateAccount(w, addr2, so2, true))
	erigonAcc2, err := state.stateReader.ReadAccountData(ecommon.Address(addr2))
	require.NoError(t, err)
	expOperaAcc2 := transformErigonAccount(erigonAcc2)
	compareAccounts(&so2.data, &expOperaAcc2, t)

	// test CreateContract
	operaAcc3, addr3 := randomStateAccount(t, false)
	operaAcc3.Nonce = 2212
	state.CreateAccount(addr3, true)
	so3 := state.getStateObject(addr3)
	require.True(t, so3.created)
	require.NoError(t, updateAccount(w, addr3, so3, true))

	// test UpdateAccountCode
	_, addr4 := randomStateAccount(t, false)
	code := []byte{'c', 'a', 'f', 'e'}
	state.SetCode(addr4, code)
	so4 := state.getStateObject(addr4)
	require.NotNil(t, so4.code)
	require.True(t, so4.dirtyCode)
	require.NoError(t, updateAccount(w, addr4, so4, true))
	codeHash := crypto.Keccak256Hash(code)
	expCode, err := state.stateReader.ReadAccountCode(ecommon.Address(addr4), so4.data.Incarnation, ecommon.Hash(codeHash))
	require.NoError(t, err)
	require.NotNil(t, expCode)
	require.True(t, bytes.Equal(expCode, code))
}

func compareAccounts(acc1, acc2 *Account, t *testing.T) {
	if acc1.Balance.Cmp(acc2.Balance) != 0 {
		t.Fatalf("Balance mismatch: have %v, want %v", acc1.Balance, acc2.Balance)
	}
	if acc1.Nonce != acc2.Nonce {
		t.Fatalf("Nonce mismatch: have %v, want %v", acc1.Nonce, acc2.Nonce)
	}
	if acc1.Root != acc2.Root {
		t.Errorf("Root mismatch: have %x, want %x", acc1.Root[:], acc2.Root[:])
	}
	if !bytes.Equal(acc1.CodeHash, acc2.CodeHash) {
		t.Fatalf("CodeHash mismatch: have %v, want %v", acc1.CodeHash, acc2.CodeHash)
	}

	if acc1.Incarnation != acc2.Incarnation {
		t.Fatalf("CodeHash mismatch: have %v, want %v", acc1.Incarnation, acc2.Incarnation)
	}

}

func randomStateAccount(t *testing.T, isContract bool) (*Account, common.Address) {
	t.Helper()
	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	addr := crypto.PubkeyToAddress(key.PublicKey)
	acc := newAccount()
	acc.Balance = big.NewInt(0).Set(big.NewInt(rand.Int63()))
	acc.Root = common.HexToHash("dsd9sd9302rf3ug9f3r0gir0gfir0egfidgf9")
	acc.CodeHash = []byte("ssdksdksdkslkdmskfdms3934r09fjid09eikd09eid")
	if isContract {
		acc.Incarnation = 1
	}

	return acc, addr
}
