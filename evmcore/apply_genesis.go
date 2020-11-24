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
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/log"

	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/opera"
	"github.com/Fantom-foundation/go-opera/opera/genesis"
)

// ApplyGenesis writes or updates the genesis block in db.
func ApplyGenesis(db ethdb.Database, g opera.GenesisState) (*EvmBlock, error) {
	// state
	statedb, err := state.New(common.Hash{}, state.NewDatabase(db), nil)
	if err != nil {
		return nil, err
	}
	g.Accounts.ForEach(func(addr common.Address, account genesis.Account) {
		statedb.AddBalance(addr, account.Balance)
		statedb.SetCode(addr, account.Code)
		statedb.SetNonce(addr, account.Nonce)
	})
	g.Storage.ForEach(func(addr common.Address, key common.Hash, value common.Hash) {
		statedb.SetState(addr, key, value)
	})

	// initial block
	root, err := statedb.Commit(true)
	if err != nil {
		return nil, err
	}
	block := genesisBlock(g, root)

	err = statedb.Database().TrieDB().Cap(0)
	if err != nil {
		return nil, err
	}

	return block, nil
}

// genesisBlock makes genesis block with pretty hash.
func genesisBlock(g opera.GenesisState, root common.Hash) *EvmBlock {

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
func MustApplyGenesis(g opera.GenesisState, db ethdb.Database) *EvmBlock {
	block, err := ApplyGenesis(db, g)
	if err != nil {
		log.Crit("ApplyGenesis", "err", err)
	}
	return block
}
