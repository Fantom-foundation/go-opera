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
	"context"
	"crypto/ecdsa"
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"time"

	"github.com/Fantom-foundation/go-opera/erigon"
	"github.com/Fantom-foundation/go-opera/gossip/evmstore/state"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"

	"github.com/Fantom-foundation/go-opera/inter"

	"github.com/ledgerwatch/erigon-lib/kv"

	estate "github.com/ledgerwatch/erigon/core/state"
)

var FakeGenesisTime = inter.Timestamp(1608600000 * time.Second)

// ApplyFakeGenesis writes or updates the genesis block in db.
func ApplyFakeGenesis(db kv.RwDB, time inter.Timestamp, balances map[common.Address]*big.Int) (*EvmBlock, error) {

	tx, err := db.BeginRw(context.Background())
	if err != nil {
		panic(err)
	}

	defer tx.Rollback()

	statedb := state.NewWithStateReader(estate.NewPlainStateReader(tx))

	for acc, balance := range balances {
		statedb.SetBalance(acc, balance)
	}

	if err := statedb.CommitBlock(estate.NewPlainStateWriterNoHistory(tx)); err != nil {
		return nil, err
	}

	if err := tx.ClearBucket(kv.HashedAccounts); err != nil {
		return nil, fmt.Errorf("clear HashedAccounts bucket: %w", err)
	}
	if err := tx.ClearBucket(kv.HashedStorage); err != nil {
		return nil, fmt.Errorf("clear HashedStorage bucket: %w", err)
	}
	if err := tx.ClearBucket(kv.TrieOfAccounts); err != nil {
		return nil, fmt.Errorf("clear TrieOfAccounts bucket: %w", err)
	}
	if err := tx.ClearBucket(kv.TrieOfStorage); err != nil {
		return nil, fmt.Errorf("clear TrieOfStorage bucket: %w", err)
	}

	if err := erigon.GenerateHashedStatePut(tx); err != nil {
		return nil, err
	}

	// initial block
	root, err := erigon.CalcRoot("FakeGenesis", tx)
	if err != nil {
		return nil, err
	}

	tx.Commit()

	block := genesisBlock(time, common.Hash(root))

	return block, nil
}

// genesisBlock makes genesis block with pretty hash.
func genesisBlock(time inter.Timestamp, root common.Hash) *EvmBlock {
	block := &EvmBlock{
		EvmHeader: EvmHeader{
			Number:   big.NewInt(0),
			Time:     time,
			GasLimit: math.MaxUint64,
			Root:     root,
			TxHash:   types.EmptyRootHash,
		},
	}

	return block
}

// MustApplyFakeGenesis writes the genesis block and state to db, panicking on error.
func MustApplyFakeGenesis(db kv.RwDB, time inter.Timestamp, balances map[common.Address]*big.Int) *EvmBlock {
	block, err := ApplyFakeGenesis(db, time, balances)
	if err != nil {
		log.Crit("ApplyFakeGenesis", "err", err)
	}
	return block
}

// FakeKey gets n-th fake private key.
func FakeKey(n int) *ecdsa.PrivateKey {
	reader := rand.New(rand.NewSource(int64(n)))

	key, err := ecdsa.GenerateKey(crypto.S256(), reader)
	if err != nil {
		panic(err)
	}

	return key
}
