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
	"math/rand"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"

	"github.com/Fantom-foundation/go-opera/inter"
)

var FakeGenesisTime = inter.Timestamp(1608600000 * time.Second)

// ApplyFakeGenesis writes or updates the genesis block in db.
func ApplyFakeGenesis(statedb StateDB, time inter.Timestamp, balances map[common.Address]*big.Int) (*EvmBlock, error) {
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

func flush(statedb StateDB, clean bool) (root common.Hash, err error) {
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
func MustApplyFakeGenesis(statedb StateDB, time inter.Timestamp, balances map[common.Address]*big.Int) *EvmBlock {
	block, err := ApplyFakeGenesis(statedb, time, balances)
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
