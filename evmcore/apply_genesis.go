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
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/log"

	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/opera"
	"github.com/Fantom-foundation/go-opera/opera/genesis"
)

// ApplyGenesis writes or updates the genesis block in db.
func ApplyGenesis(statedb *state.StateDB, g opera.Genesis, maxMemoryUsage int) (*EvmBlock, error) {
	mem := 0
	g.Accounts.ForEach(func(addr common.Address, account genesis.Account) {
		statedb.AddBalance(addr, account.Balance)
		statedb.SetCode(addr, account.Code)
		statedb.SetNonce(addr, account.Nonce)
		mem += 1024 + len(account.Code)
		if mem > maxMemoryUsage {
			_, _ = flush(statedb, true)
			mem = 0
		}
	})
	g.Storage.ForEach(func(addr common.Address, key common.Hash, value common.Hash) {
		statedb.SetState(addr, key, value)
		mem += 1024
		if mem > maxMemoryUsage {
			_, _ = flush(statedb, true)
			mem = 0
		}
	})

	// initial block
	root, err := flush(statedb, true)
	if err != nil {
		return nil, err
	}
	block := genesisBlock(g, root)

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
	return
}

// genesisBlock makes genesis block with pretty hash.
func genesisBlock(g opera.Genesis, root common.Hash) *EvmBlock {
	block := &EvmBlock{
		EvmHeader: EvmHeader{
			Number:   big.NewInt(0),
			Time:     g.Time,
			GasLimit: math.MaxUint64,
			Root:     root,
			TxHash:   common.Hash(inter.EmptyTxHash),
		},
	}

	return block
}

// MustApplyGenesis writes the genesis block and state to db, panicking on error.
func MustApplyGenesis(g opera.Genesis, statedb *state.StateDB, maxMemoryUsage int) *EvmBlock {
	block, err := ApplyGenesis(statedb, g, maxMemoryUsage)
	if err != nil {
		log.Crit("ApplyGenesis", "err", err)
	}
	return block
}
