// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package state

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/stretchr/testify/suite"

	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/erigon-lib/kv/memdb"

	ecommon "github.com/ledgerwatch/erigon/common"
	estate "github.com/ledgerwatch/erigon/core/state"
)

type IntegrationTestSuite struct {
	suite.Suite

	kv    kv.RwDB
	tx    kv.RwTx
	state *StateDB
	r     estate.StateReader
	w     estate.StateWriter
}

func (s *IntegrationTestSuite) SetupTest() {
	_, s.tx = memdb.NewTestTx(s.T())

	s.r = estate.NewPlainStateReader(s.tx)
	s.w = estate.NewPlainStateWriterNoHistory(s.tx)
	s.state = NewWithStateReader(s.r)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.tx.Rollback()
}

//TODO fix it:  add root of storage trie
/*
func (s *IntegrationTestSuite) TestDump() {
	// generate a few entries
	obj1 := s.state.GetOrNewStateObject(common.BytesToAddress([]byte{0x01}))
	obj1.AddBalance(big.NewInt(22))
	obj2 := s.state.GetOrNewStateObject(common.BytesToAddress([]byte{0x01, 0x02}))
	obj2.SetCode(crypto.Keccak256Hash([]byte{3, 3, 3, 3, 3, 3, 3}), []byte{3, 3, 3, 3, 3, 3, 3})
	obj3 := s.state.GetOrNewStateObject(common.BytesToAddress([]byte{0x02}))
	obj3.SetBalance(big.NewInt(44))

	eAcc1 := transformStateAccount(&obj1.data, false)
	eAcc2 := transformStateAccount(&obj2.data, false)

	// write some of them to the trie
	err := s.w.UpdateAccountData(ecommon.Address(obj1.address), &eAcc1, &eAcc1)
	s.Require().NoError(err)
	err = s.w.UpdateAccountData(ecommon.Address(obj2.address), &eAcc2, &eAcc2)
	s.Require().NoError(err)

	err = s.state.CommitBlock(s.w)
	s.Require().NoError(err)

	// check that dump contains the state objects that are in trie

	// check that DumpToCollector contains the state objects that are in trie
	got := string(s.state.Dump(nil, s.tx))
	s.T().Log("got", got)
	s.Require().Equal(2,3)
	want := `{
    "root": "71edff0130dd2385947095001c73d9e28d862fc286fca2b922ca6f6f3cddfdd2",
    "accounts": {
        "0x0000000000000000000000000000000000000001": {
            "balance": "22",
            "nonce": 0,
            "root": "0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421",
            "codeHash": "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470",
            "key": "0x1468288056310c82aa4c01a7e12a10f8111a0560e72b700555479031b86c357d"
        },
        "0x0000000000000000000000000000000000000002": {
            "balance": "44",
            "nonce": 0,
            "root": "0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421",
            "codeHash": "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470",
            "key": "0xd52688a8f926c816ca1e079067caba944f158e764817b83fc43594370ca9cf62"
        },
        "0x0000000000000000000000000000000000000102": {
            "balance": "0",
            "nonce": 0,
            "root": "0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421",
            "codeHash": "0x87874902497a5bb968da31a2998d8f22e949d1ef6214bcdedd8bae24cca4b9e3",
            "code": "0x03030303030303",
            "key": "0xa17eacbc25cda025e81db9c5c62868822c73ce097cee2a63e33a2e41268358a1"
        }
    }
}`
	if got != want {
		s.T().Errorf("DumpToCollector mismatch:\ngot: %s\nwant: %s\n", got, want)
	}
}
*/

func (s *IntegrationTestSuite) TestNull() {
	address := common.HexToAddress("0x823140710bf13990e4500136726d8b55")
	s.state.CreateAccount(address, true)
	//value := common.FromHex("0x823140710bf13990e4500136726d8b55")
	value := common.Hash{}

	s.state.SetState(address, common.Hash{}, value)

	err := s.state.CommitBlock(s.w)
	s.Require().NoError(err)

	value = s.state.GetCommittedState(address, common.Hash{})
	s.Require().Equal(value, common.Hash{}, "expected empty hash. got %x", value)
}

func (s *IntegrationTestSuite) TestTouchDelete() {
	s.state.GetOrNewStateObject(common.Address{})

	err := s.state.CommitBlock(s.w)
	s.Require().NoError(err)

	s.state.Reset()

	snapshot := s.state.Snapshot()
	s.state.AddBalance(common.Address{}, big.NewInt(0))

	s.Require().Len(s.state.journal.dirties, 1)
	s.state.RevertToSnapshot(snapshot)
	s.Require().Len(s.state.journal.dirties, 0)
}

func (s *IntegrationTestSuite) TestStateObjectWithNilAccounts() {
	addr := common.BytesToAddress([]byte("aa"))

	_, ok := s.state.nilAccounts[addr]
	s.Require().False(ok)
	so := s.state.getStateObject(addr)
	_, ok = s.state.nilAccounts[addr]
	s.Require().True(ok)

	so1, exist := s.state.stateObjects[addr]
	s.Require().True(exist)
	compareStateObjects(so, so1, s.T())
}

func (s *IntegrationTestSuite) TestCreateAccount() {

	addr := common.BytesToAddress([]byte("aa"))
	so := s.state.getStateObject(addr)
	so.setBalance(big.NewInt(232323))

	inc, err := s.state.stateReader.ReadAccountIncarnation(ecommon.Address(addr))
	s.Require().NoError(err)
	s.Require().Equal(inc, uint64(0))

	erigonAcc := transformStateAccount(&so.data, true)
	s.Require().NoError(s.w.UpdateAccountData(ecommon.Address(addr), nil, &erigonAcc))
	s.Require().NoError(s.w.DeleteAccount(ecommon.Address(addr), &erigonAcc))

	inc, err = s.state.stateReader.ReadAccountIncarnation(ecommon.Address(addr))
	s.Require().NoError(err)
	s.Require().Equal(inc, uint64(1))

	s.state.CreateAccount(addr, true)

	so1 := s.state.getStateObject(addr)
	s.Require().True(so1.created)
	s.Require().Equal(so1.data.Incarnation, uint64(2))
	s.Require().Equal(so1.Balance().Cmp(so.Balance()), 0)

	addr2 := common.BytesToAddress([]byte("bb"))
	s.state.SetNonce(addr2, 50)
	so2 := s.state.getStateObject(addr2)
	erigonAcc2 := transformStateAccount(&so2.data, false)

	inc, err = s.state.stateReader.ReadAccountIncarnation(ecommon.Address(addr2))
	s.Require().NoError(err)
	s.Require().Equal(inc, uint64(0))

	s.Require().NoError(s.w.UpdateAccountData(ecommon.Address(addr2), nil, &erigonAcc2))
	s.Require().NoError(s.w.DeleteAccount(ecommon.Address(addr2), &erigonAcc2))

	s.state.CreateAccount(addr, false)

	so3 := s.state.getStateObject(addr2)
	s.Require().False(so3.created)
	s.Require().Equal(so3.data.Incarnation, uint64(0))
	s.Require().Equal(so2.Balance().Cmp(so3.Balance()), 0)
}

func (s *IntegrationTestSuite) TestUpdateReadAccountData() {
	addr := common.BytesToAddress([]byte("aa"))
	so := s.state.getStateObject(addr)
	so.setBalance(big.NewInt(232323))
	erigonAcc := transformStateAccount(&so.data, false)
	s.Require().NoError(s.w.UpdateAccountData(ecommon.Address(addr), nil, &erigonAcc))
	expErigonAcc, err := s.state.stateReader.ReadAccountData(ecommon.Address(addr))
	s.Require().NoError(err)
	s.Require().NotNil(expErigonAcc)
	expOperaAcc := transformErigonAccount(expErigonAcc)
	compareAccounts(&so.data, &expOperaAcc, s.T())

	addr1 := common.BytesToAddress([]byte("bb"))
	so1 := s.state.getStateObject(addr1)
	erigonAcc1 := transformStateAccount(&so1.data, false)
	s.Require().NoError(s.w.UpdateAccountData(ecommon.Address(addr), nil, &erigonAcc1))
	s.Require().NoError(s.w.DeleteAccount(ecommon.Address(addr1), &erigonAcc1))
	expErigonAcc1, err := s.state.stateReader.ReadAccountData(ecommon.Address(addr1))
	s.Require().NoError(err)
	s.Require().Nil(expErigonAcc1)
}

// add couple more cases
func (s *IntegrationTestSuite) TestUpdateAccountCode() {

	acc, addr := randomStateAccount(s.T(), true)
	erigonAcc := transformStateAccount(acc, true)
	s.Require().NoError(s.w.UpdateAccountCode(ecommon.Address(addr), erigonAcc.Incarnation, erigonAcc.CodeHash, acc.CodeHash))
	code, err := s.state.stateReader.ReadAccountCode(ecommon.Address(addr), erigonAcc.Incarnation, erigonAcc.CodeHash)
	s.Require().NoError(err)
	s.Require().NotNil(code)
	s.Require().True(bytes.Equal(code, acc.CodeHash))

	acc1, addr1 := randomStateAccount(s.T(), true)
	erigonAcc1 := transformStateAccount(acc1, true)
	s.Require().NoError(s.w.UpdateAccountCode(ecommon.Address(addr1), erigonAcc1.Incarnation, erigonAcc1.CodeHash, nil))
	code1, err := s.state.stateReader.ReadAccountCode(ecommon.Address(addr1), erigonAcc1.Incarnation, erigonAcc.CodeHash)
	s.Require().NoError(err)
	s.Require().Equal(code1, []byte{0x0})

	acc2, addr2 := randomStateAccount(s.T(), true)
	erigonAcc2 := transformStateAccount(acc2, true)
	s.Require().NoError(s.w.UpdateAccountCode(ecommon.Address(addr2), erigonAcc2.Incarnation, erigonAcc2.CodeHash, nil))
	code2, err := s.state.stateReader.ReadAccountCode(ecommon.Address(addr1), erigonAcc1.Incarnation, ecommon.HexToHash("random"))
	s.Require().NoError(err)
	s.Require().Nil(code2)
}

// test UpdateAccountStorage
func (s *IntegrationTestSuite) TestStateObjectUpdateTrie() {
	contract := common.HexToAddress("0x71dd1027069078091B3ca48093B00E4735B20624")
	storageKey := common.HexToHash("0x0e4c0e7175f9d22279a4f63ff74f7fa28b7a954a6454debaa62ce43dd9132541")
	storageValue := common.HexToHash("0x016345785d8a0000")

	s.state.SetState(contract, storageKey, storageValue)
	so := s.state.getStateObject(contract)
	s.Require().Len(so.dirtyStorage, 1)
	s.Require().Len(so.originStorage, 0)

	s.Require().NoError(so.updateTrie(s.w, so.data.Incarnation))

	so = s.state.getStateObject(contract)
	s.Require().Len(so.dirtyStorage, 1)
	s.Require().Len(so.originStorage, 1)

	eKey := ecommon.Hash(storageKey)
	storageByteValue, err := s.state.stateReader.ReadAccountStorage(ecommon.Address(contract), so.data.Incarnation, &eKey)
	s.Require().NoError(err)
	expStorageValue := common.BytesToHash(storageByteValue)
	s.Require().Equal(storageValue, expStorageValue)

	// set invalidInc while ReadAccountStorage
	contract1 := common.HexToAddress("0x71dd1027069078091B3ca48093B00E4735B20625")
	storageKey1 := common.HexToHash("0x0e4c0e7175f9d22279a4f63ff74f7fa28b7a954a6454debaa62ce43dd9132542")
	storageValue1 := common.HexToHash("0x016345785d8a00001")

	s.state.SetState(contract1, storageKey1, storageValue1)
	so = s.state.getStateObject(contract1)
	s.Require().Len(so.dirtyStorage, 1)
	s.Require().Len(so.originStorage, 0)

	s.Require().NoError(so.updateTrie(s.w, so.data.Incarnation))

	so = s.state.getStateObject(contract1)
	s.Require().Len(so.dirtyStorage, 1)
	s.Require().Len(so.originStorage, 1)

	eKey1 := ecommon.Hash(storageKey1)
	invalidInc := so.data.Incarnation + 1
	expStorageValue1, err := s.state.stateReader.ReadAccountStorage(ecommon.Address(contract1), invalidInc, &eKey1)
	s.Require().Nil(expStorageValue1)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TestSnapshot() {
	stateobjaddr := common.BytesToAddress([]byte("aa"))
	var storageaddr common.Hash
	data1 := common.BytesToHash([]byte{42})
	data2 := common.BytesToHash([]byte{43})

	// snapshot the genesis state
	genesis := s.state.Snapshot()

	// set initial state object value
	s.state.SetState(stateobjaddr, storageaddr, data1)
	snapshot := s.state.Snapshot()

	// set a new state object value, revert it and ensure correct content
	s.state.SetState(stateobjaddr, storageaddr, data2)
	s.state.RevertToSnapshot(snapshot)

	if v := s.state.GetState(stateobjaddr, storageaddr); v != data1 {
		s.T().Errorf("wrong storage value %v, want %v", v, data1)
	}
	if v := s.state.GetCommittedState(stateobjaddr, storageaddr); v != (common.Hash{}) {
		s.T().Errorf("wrong committed storage value %v, want %v", v, common.Hash{})
	}

	// revert up to the genesis state and ensure correct content
	s.state.RevertToSnapshot(genesis)
	if v := s.state.GetState(stateobjaddr, storageaddr); v != (common.Hash{}) {
		s.T().Errorf("wrong storage value %v, want %v", v, common.Hash{})
	}
	if v := s.state.GetCommittedState(stateobjaddr, storageaddr); v != (common.Hash{}) {
		s.T().Errorf("wrong committed storage value %v, want %v", v, common.Hash{})
	}
}

func TestSnapshot2(t *testing.T) {
	_, tx := memdb.NewTestTx(t)
	w := estate.NewPlainState(tx, 1)
	state := NewWithStateReader(estate.NewPlainState(tx, 1))

	stateobjaddr0 := common.BytesToAddress([]byte("so0"))
	stateobjaddr1 := common.BytesToAddress([]byte("so1"))
	var storageaddr common.Hash

	data0 := common.BytesToHash([]byte{17})
	data1 := common.BytesToHash([]byte{18})

	state.SetState(stateobjaddr0, storageaddr, data0)
	state.SetState(stateobjaddr1, storageaddr, data1)

	// db  are already non-empty values
	t.Log("state.getStateObject")
	so0 := state.getStateObject(stateobjaddr0)

	so0.SetBalance(big.NewInt(42))
	so0.SetNonce(43)
	so0.SetCode(crypto.Keccak256Hash([]byte{'c', 'a', 'f', 'e'}), []byte{'c', 'a', 'f', 'e'})
	so0.suicided = false
	so0.deleted = false
	state.setStateObject(so0)

	err := state.CommitBlock(w)
	if err != nil {
		t.Fatal("error while commting a state", err)
	}

	// and one with deleted == true
	so1 := state.getStateObject(stateobjaddr1)
	so1.SetBalance(big.NewInt(52))
	so1.SetNonce(53)
	so1.SetCode(crypto.Keccak256Hash([]byte{'c', 'a', 'f', 'e', '2'}), []byte{'c', 'a', 'f', 'e', '2'})
	so1.suicided = true
	so1.deleted = true
	state.setStateObject(so1)

	so1 = state.getStateObject(stateobjaddr1)
	if so1 != nil && !so1.deleted {
		t.Fatalf("deleted object not nil when getting")
	}

	snapshot := state.Snapshot()
	state.RevertToSnapshot(snapshot)

	so0Restored := state.getStateObject(stateobjaddr0)
	if so0Restored == nil {
		t.Fatal("so0Restored is nil")
	}

	// Update lazily-loaded values before comparing.
	so0Restored.GetState(storageaddr)
	so0Restored.Code()
	// non-deleted is equal (restored)
	compareStateObjects(so0Restored, so0, t)

	// deleted should be nil, both before and after restore of state copy
	so1Restored := state.getStateObject(stateobjaddr1)
	if so1Restored != nil && !so1Restored.deleted {
		t.Fatalf("deleted object not nil after restoring snapshot: %+v", so1Restored)
	}
}

func compareStateObjects(so0, so1 *stateObject, t *testing.T) {
	if so0.Address() != so1.Address() {
		t.Fatalf("Address mismatch: have %v, want %v", so0.address, so1.address)
	}
	if so0.Balance().Cmp(so1.Balance()) != 0 {
		t.Fatalf("Balance mismatch: have %v, want %v", so0.Balance(), so1.Balance())
	}
	if so0.Nonce() != so1.Nonce() {
		t.Fatalf("Nonce mismatch: have %v, want %v", so0.Nonce(), so1.Nonce())
	}
	if so0.data.Root != so1.data.Root {
		t.Errorf("Root mismatch: have %x, want %x", so0.data.Root[:], so1.data.Root[:])
	}
	if !bytes.Equal(so0.CodeHash(), so1.CodeHash()) {
		t.Fatalf("CodeHash mismatch: have %v, want %v", so0.CodeHash(), so1.CodeHash())
	}
	if !bytes.Equal(so0.code, so1.code) {
		t.Fatalf("Code mismatch: have %v, want %v", so0.code, so1.code)
	}

	if len(so1.dirtyStorage) != len(so0.dirtyStorage) {
		t.Errorf("Dirty storage size mismatch: have %d, want %d", len(so1.dirtyStorage), len(so0.dirtyStorage))
	}
	for k, v := range so1.dirtyStorage {
		if so0.dirtyStorage[k] != v {
			t.Errorf("Dirty storage key %x mismatch: have %v, want %v", k, so0.dirtyStorage[k], v)
		}
	}
	for k, v := range so0.dirtyStorage {
		if so1.dirtyStorage[k] != v {
			t.Errorf("Dirty storage key %x mismatch: have %v, want none.", k, v)
		}
	}
	if len(so1.originStorage) != len(so0.originStorage) {
		t.Errorf("Origin storage size mismatch: have %d, want %d", len(so1.originStorage), len(so0.originStorage))
	}
	for k, v := range so1.originStorage {
		if so0.originStorage[k] != v {
			t.Errorf("Origin storage key %x mismatch: have %v, want %v", k, so0.originStorage[k], v)
		}
	}
	for k, v := range so0.originStorage {
		if so1.originStorage[k] != v {
			t.Errorf("Origin storage key %x mismatch: have %v, want none.", k, v)
		}
	}
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
