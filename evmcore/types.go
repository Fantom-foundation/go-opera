// Copyright 2015 The go-ethereum Authors
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

package evmcore

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
)

// StateDB is an EVM database for full state querying.
type StateDB interface {
	vm.StateDB

	SetBalance(addr common.Address, amount *big.Int)
	// Database retrieves the low level database supporting the lower level trie ops.
	Database() state.Database
	// Prepare sets the current transaction hash and index which are
	// used when the EVM emits new state logs.
	Prepare(thash common.Hash, ti int)
	// TxIndex returns the current transaction index set by Prepare.
	TxIndex() int
	GetLogs(hash common.Hash, blockHash common.Hash) []*types.Log
	// IntermediateRoot computes the current root hash of the state trie.
	// It is called in between transactions to get the root hash that
	// goes into transaction receipts.
	IntermediateRoot(deleteEmptyObjects bool) common.Hash
	// Finalise finalises the state by removing the s destructed objects and clears
	// the journal as well as the refunds. Finalise, however, will not push any updates
	// into the tries just yet. Only IntermediateRoot or Commit will do that.
	Finalise(deleteEmptyObjects bool)
	// Commit writes the state to the underlying in-memory trie database.
	Commit(deleteEmptyObjects bool) (common.Hash, error)

	Error() error
	Copy() StateDB
	SetStorage(addr common.Address, storage map[common.Hash]common.Hash)

	Measurements() Measurements
}

// Measurements gathered during execution for debugging purposes
type Measurements struct {
	AccountReads         time.Duration
	AccountHashes        time.Duration
	AccountUpdates       time.Duration
	AccountCommits       time.Duration
	StorageReads         time.Duration
	StorageHashes        time.Duration
	StorageUpdates       time.Duration
	StorageCommits       time.Duration
	SnapshotAccountReads time.Duration
	SnapshotStorageReads time.Duration
	SnapshotCommits      time.Duration
}

// Validator is an interface which defines the standard for block validation. It
// is only responsible for validating block contents, as the header validation is
// done by the specific consensus engines.
type Validator interface {
	// ValidateBody validates the given block's content.
	ValidateBody(block *EvmBlock) error

	// ValidateState validates the given statedb and optionally the receipts and
	// gas used.
	ValidateState(block *EvmBlock, state StateDB, receipts types.Receipts, usedGas uint64) error
}

// Prefetcher is an interface for pre-caching transaction signatures and state.
type Prefetcher interface {
	// Prefetch processes the state changes according to the Ethereum rules by running
	// the transaction messages using the statedb, but any changes are discarded. The
	// only goal is to pre-cache transaction signatures and state trie nodes.
	Prefetch(block *EvmBlock, statedb StateDB, cfg vm.Config, interrupt *uint32)
}

// Processor is an interface for processing blocks using a given initial state.
type Processor interface {
	// Process processes the state changes according to the Ethereum rules by running
	// the transaction messages using the statedb and applying any rewards to both
	// the processor (coinbase) and any included uncles.
	Process(block *EvmBlock, statedb StateDB, cfg vm.Config) (types.Receipts, []*types.Log, uint64, error)
}

func ToStateDB(statedb *state.StateDB) StateDB {
	return &stateWrapper{statedb}
}

func IsMptStateDB(st StateDB) (statedb *state.StateDB, ok bool) {
	if wrapper, is := st.(*stateWrapper); is {
		statedb = wrapper.StateDB
		ok = true
		return
	}
	return
}

// stateWrapper casts *state.StateDB to StateDB interface.
type stateWrapper struct {
	*state.StateDB
}

func (w *stateWrapper) Copy() StateDB {
	statedb := w.StateDB.Copy()
	return &stateWrapper{statedb}
}

func (w *stateWrapper) Measurements() Measurements {
	return Measurements{
		AccountReads:   w.StateDB.AccountReads,
		AccountHashes:  w.StateDB.AccountHashes,
		AccountUpdates: w.StateDB.AccountUpdates,
		AccountCommits: w.StateDB.AccountCommits,
		StorageReads:   w.StateDB.StorageReads,
		StorageHashes:  w.StateDB.StorageHashes,
		StorageUpdates: w.StateDB.StorageUpdates,
		StorageCommits: w.StateDB.StorageCommits,
	}
}
