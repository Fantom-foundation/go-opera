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
	"crypto/ecdsa"
	"math"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"

	"github.com/Fantom-foundation/go-opera/inter"
)

var FakeGenesisTime = inter.Timestamp(1608600000 * time.Second)

var key0, _ = crypto.ToECDSA(hexutil.MustDecode("0x96b6d2f6d6b677f4094f4f9c759f1d7b01ab8bff9660b8e4828687ac4ccf4652"))
var key1, _ = crypto.ToECDSA(hexutil.MustDecode("0xf080fbe021483951604c735d0bae7999c542912a7065abd58052f5206d5db23c"))
var key2, _ = crypto.ToECDSA(hexutil.MustDecode("0x5f546197c7bccfb5357a62349418e176b41d806980f6a29a666cc1b25d83c5e9"))
var key3, _ = crypto.ToECDSA(hexutil.MustDecode("0x6df728689a875aa47aa6e4c8369ed134dd0fc5bc628842e4733d64f93208ffec"))
var key4, _ = crypto.ToECDSA(hexutil.MustDecode("0xbdfc91fd521f14cae1f2c2087e797188a9852c1a6141e82b95f531a8a33d26f2"))
var key5, _ = crypto.ToECDSA(hexutil.MustDecode("0xb5ab5968e871286d309d0880165b473f8fd7a783d3a82aa42ce2f5ae3e4a5d30"))
var key6, _ = crypto.ToECDSA(hexutil.MustDecode("0x20afac5edface492fa9f9a2b908b26c5e7fc6c6f0e091c01a7809e39562a3bff"))
var key7, _ = crypto.ToECDSA(hexutil.MustDecode("0xeb985bb616c93db9626030788e520d7fdbf0135fb9610c9b8d3c29c75dd3e75d"))
var key8, _ = crypto.ToECDSA(hexutil.MustDecode("0xd4a201b1f2542352e47ec1508ba0c48c7a08e5667657ec416c3effa235b76c0b"))
var key9, _ = crypto.ToECDSA(hexutil.MustDecode("0xdd3ca9fae341080f4a3a13792c105241b5720a3bfd058cb61b27a1a82c187bf3"))

var keys = [11]*ecdsa.PrivateKey{key0, key1, key2, key3, key4, key5, key6, key7, key8, key9, key0}

// ApplyFakeGenesis writes or updates the genesis block in db.
func ApplyFakeGenesis(statedb *state.StateDB, time inter.Timestamp, balances map[common.Address]*big.Int) (*EvmBlock, error) {
	for acc, balance := range balances {
		statedb.SetBalance(acc, balance)
	}

	// initial block
	root, err := flush(statedb, true)
	if err != nil {
		return nil, err
	}
	block := genesisBlock(time, root)

	return block, nil
}

func flush(statedb *state.StateDB, clean bool) (root common.Hash, err error) {
	root, err = statedb.Commit(clean)
	if err != nil {
		return
	}
	err = statedb.Database().TrieDB().Commit(root, false, nil)
	if err != nil {
		return
	}

	if !clean {
		err = statedb.Database().TrieDB().Cap(0)
	}

	return
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
func MustApplyFakeGenesis(statedb *state.StateDB, time inter.Timestamp, balances map[common.Address]*big.Int) *EvmBlock {
	block, err := ApplyFakeGenesis(statedb, time, balances)
	if err != nil {
		log.Crit("ApplyFakeGenesis", "err", err)
	}
	return block
}

// FakeKey gets n-th fake private key.
func FakeKey(n int) *ecdsa.PrivateKey {
	return keys[n]
}
