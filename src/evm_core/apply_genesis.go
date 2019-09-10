package evm_core

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"

	"github.com/Fantom-foundation/go-lachesis/src/lachesis/genesis"
)

// GenesisMismatchError is raised when trying to overwrite an existing
// genesis block with an incompatible one.
type GenesisMismatchError struct {
	Stored, New common.Hash
}

func (e *GenesisMismatchError) Error() string {
	return fmt.Sprintf("database contains incompatible genesis (have %x, new %x)", e.Stored, e.New)
}

// ApplyGenesis writes or updates the genesis block in db.
func ApplyGenesis(db ethdb.Database, genesis *genesis.Genesis) (*EvmBlock, error) {
	if genesis == nil {
		return nil, ErrNoGenesis
	}

	// state
	statedb, err := state.New(common.Hash{}, state.NewDatabase(db))
	if err != nil {
		return nil, err
	}
	for addr, account := range genesis.Alloc {
		statedb.AddBalance(addr, account.Balance)
		statedb.SetCode(addr, account.Code)
		statedb.SetNonce(addr, account.Nonce)
		for key, value := range account.Storage {
			statedb.SetState(addr, key, value)
		}
	}

	// initial block
	block := &EvmBlock{
		EvmHeader: EvmHeader{
			Number:   big.NewInt(0),
			Time:     genesis.Time,
			GasLimit: params.GenesisGasLimit, // TODO: config
			Coinbase: common.BytesToAddress([]byte{1}),
			Root:     statedb.IntermediateRoot(false),
		},
	}
	block.Hash = block.PrettyHash()
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

	root, err := statedb.Commit(false)
	if err != nil {
		return nil, err
	}
	err = statedb.Database().TrieDB().Commit(root, true) //TODO: ???
	if err != nil {
		return nil, err
	}

	writeBlockIndexes(db, blockNum, block.Hash)

	return block, nil
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
func mustApplyGenesis(g *genesis.Genesis, db ethdb.Database) *EvmBlock {
	block, err := ApplyGenesis(db, g)
	if err != nil {
		log.Crit("ApplyGenesis", "err", err)
	}
	return block
}
