package evm_core

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
// Returns genesis FiWitness hash, StateHash
func ApplyGenesis(db ethdb.Database, genesis *genesis.Genesis, genesisHashFn func(*EvmHeader) common.Hash) (common.Hash, common.Hash, error) {
	if genesis == nil {
		return common.Hash{}, common.Hash{}, ErrNoGenesis
	}
	b := genesisToBlock(nil, genesis, genesisHashFn)
	stored := rawdb.ReadCanonicalHash(db, 0)
	if (stored == common.Hash{}) {
		// Just commit the new block if there is no stored genesis block.
		log.Info("Writing genesis state")
		block, err := genesisWrite(db, genesis, genesisHashFn)
		if err != nil {
			return block.Hash, block.Root, err
		}
		return block.Hash, block.Root, nil
	} else if b.Hash != stored {
		// Check whether the genesis block is already written.
		return b.Hash, b.Root, &GenesisMismatchError{stored, b.Hash}
	}

	// We have the genesis block in database(perhaps in ancient database)
	// but the corresponding state is missing.
	header := rawdb.ReadHeader(db, stored, 0)
	if _, err := state.New(header.Root, state.NewDatabaseWithCache(db, 0)); err != nil {
		_, err := genesisWrite(db, genesis, genesisHashFn)
		if err != nil {
			return b.Hash, b.Root, err
		}
		return b.Hash, b.Root, nil
	}
	return b.Hash, b.Root, nil
}

// genesisToBlock creates the genesis block and writes state of a genesis specification
// to the given database (or discards it if nil).
func genesisToBlock(db ethdb.Database, g *genesis.Genesis, genesisHashFn func(*EvmHeader) common.Hash) *EvmBlock {
	if db == nil {
		db = rawdb.NewMemoryDatabase()
	}
	// write genesis accounts
	statedb, _ := state.New(common.Hash{}, state.NewDatabase(db))
	for addr, account := range g.Alloc {
		statedb.AddBalance(addr, account.Balance)
		statedb.SetCode(addr, account.Code)
		statedb.SetNonce(addr, account.Nonce)
		for key, value := range account.Storage {
			statedb.SetState(addr, key, value)
		}
	}
	root := statedb.IntermediateRoot(false)

	coinbase := common.BytesToAddress([]byte{1})

	head := &EvmHeader{
		Number:   big.NewInt(0),
		Time:     g.Time,
		GasLimit: params.GenesisGasLimit, // TODO config
		Coinbase: coinbase,
		Root:     root,
	}
	if genesisHashFn != nil {
		head.Hash = genesisHashFn(head)
	}

	statedb.Commit(false)
	statedb.Database().TrieDB().Commit(root, true)

	return &EvmBlock{
		EvmHeader: *head,
	}
}

// genesisWrite writes the block and state of a genesis specification to the database.
// The block is committed as the canonical head block.
func genesisWrite(db ethdb.Database, g *genesis.Genesis, genesisHashFn func(*EvmHeader) common.Hash) (*EvmBlock, error) {
	block := genesisToBlock(db, g, genesisHashFn)
	if block.Number.Sign() != 0 {
		return nil, fmt.Errorf("can't commit genesis state with number > 0")
	}
	rawdb.WriteReceipts(db, block.Hash, block.NumberU64(), nil)
	rawdb.WriteCanonicalHash(db, block.Hash, block.NumberU64())
	rawdb.WriteHeadBlockHash(db, block.Hash)
	rawdb.WriteHeadHeaderHash(db, block.Hash)
	rawdb.WriteHeadFastBlockHash(db, block.Hash)
	rawdb.WriteHeader(db, &types.Header{
		Number:     block.Number,
		Root:       block.Root,
		ParentHash: block.ParentHash,
		Coinbase:   block.Coinbase,
		Time:       uint64(block.Time),
	})

	return block, nil
}

// genesisMustWrite writes the genesis block and state to db, panicking on error.
// The block is committed as the canonical head block.
func genesisMustWrite(g *genesis.Genesis, db ethdb.Database, genesisHashFn func(*EvmHeader) common.Hash) *EvmBlock {
	block, err := genesisWrite(db, g, genesisHashFn)
	if err != nil {
		panic(err)
	}
	return block
}
