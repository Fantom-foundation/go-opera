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
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"

	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/lachesis"
)

// GenesisMismatchError is raised when trying to overwrite an existing
// genesis block with an incompatible one.
type GenesisMismatchError struct {
	Stored, New common.Hash
}

// Error implements error interface.
func (e *GenesisMismatchError) Error() string {
	return fmt.Sprintf("database contains incompatible genesis (have %x, new %x)", e.Stored, e.New)
}

// ApplyGenesis writes or updates the genesis block in db.
func ApplyGenesis(db ethdb.Database, net *lachesis.Config) (*EvmBlock, error) {
	if net == nil {
		return nil, ErrNoGenesis
	}

	// state
	statedb, err := state.New(common.Hash{}, state.NewDatabase(db))
	if err != nil {
		return nil, err
	}
	for addr, account := range net.Genesis.Alloc {
		statedb.AddBalance(addr, account.Balance)
		statedb.SetCode(addr, account.Code)
		statedb.SetNonce(addr, account.Nonce)
		for key, value := range account.Storage {
			statedb.SetState(addr, key, value)
		}
	}

	// initial block
	root := statedb.IntermediateRoot(false)
	block := genesisBlock(net, root)
	blockNum := block.NumberU64()

	stored := rawdb.ReadCanonicalHash(db, blockNum)
	if (stored != common.Hash{}) {
		if stored != block.Hash {
			log.Info("Other genesis block is already written", "block", stored.String())
			return nil, &GenesisMismatchError{stored, block.Hash}
		}

		log.Info("Genesis block is already written", "block", stored.String())
		return block, nil
	}

	log.Info("Commit genesis block", "block", block.Hash.String())

	root, err = statedb.Commit(false)
	if err != nil {
		return nil, err
	}
	err = statedb.Database().TrieDB().Commit(root, true) //TODO: ???
	if err != nil {
		return nil, err
	}

	writeBlockIndexes(db, blockNum, block.Hash)
	err = statedb.Database().TrieDB().Cap(0)
	if err != nil {
		return nil, err
	}

	return block, nil
}

// genesisBlock makes genesis block with pretty hash.
func genesisBlock(net *lachesis.Config, root common.Hash) *EvmBlock {

	prettyHash := func(b *EvmBlock) common.Hash {
		e := inter.NewEvent()
		// for nice-looking ID
		e.Epoch = 0
		e.Lamport = idx.Lamport(net.Dag.EpochLen)
		// actual data hashed
		e.Extra = net.Genesis.ExtraData
		e.ClaimedTime = b.Time
		e.TxHash = b.Root
		e.Creator = b.Coinbase

		return common.Hash(e.Hash())
	}

	block := &EvmBlock{
		EvmHeader: EvmHeader{
			Number:   big.NewInt(0),
			Time:     net.Genesis.Time,
			GasLimit: params.GenesisGasLimit, // TODO: config
			Coinbase: common.BytesToAddress([]byte{1}),
			Root:     root,
			TxHash:   types.EmptyRootHash,
		},
	}
	block.Hash = prettyHash(block)

	return block
}

// writeBlockIndexes writes the block's indexes.
func writeBlockIndexes(db ethdb.Database, num uint64, hash common.Hash) {
	rawdb.WriteHeaderNumber(db, hash, num)
	rawdb.WriteReceipts(db, hash, num, nil)
	rawdb.WriteCanonicalHash(db, hash, num)
	rawdb.WriteHeadBlockHash(db, hash)
	rawdb.WriteHeadHeaderHash(db, hash)
	rawdb.WriteHeadFastBlockHash(db, hash)
}

// mustApplyGenesis writes the genesis block and state to db, panicking on error.
func mustApplyGenesis(net *lachesis.Config, db ethdb.Database) *EvmBlock {
	block, err := ApplyGenesis(db, net)
	if err != nil {
		log.Crit("ApplyGenesis", "err", err)
	}
	return block
}
