package evmstore

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	//"github.com/Fantom-foundation/lachesis-base/hash"

	"github.com/cyberbono3/go-opera/gossip/evmstore/state"
	"github.com/cyberbono3/go-opera/opera/genesis"

	"github.com/cyberbono3/go-opera/erigon"
	"github.com/cyberbono3/go-opera/gossip/evmstore/ethdb"

	"github.com/ethereum/go-ethereum/common"

	gethstate "github.com/ethereum/go-ethereum/core/state"

	"github.com/ledgerwatch/erigon-lib/kv"

	"github.com/c2h5oh/datasize"
)

const batchSizeStr = "256M"

// ApplyGenesis writes initial state.
// TODO consider to write not only Plainstate EVM accounts into EVMDB but also kv.Code kv.IncarnationMap records as well
func (s *Store) ApplyGenesis(g genesis.Genesis) (err error) {

	tx, err := s.EvmDb.RwKV().BeginRw(context.Background())
	if err != nil {
		return err
	}

	var batchSize datasize.ByteSize
	must(batchSize.UnmarshalText([]byte(batchSizeStr)))

	path := filepath.Join(erigon.DefaultDataDir(), "erigon", "batch")
	batch := ethdb.NewHashBatch(tx, path)

	defer func() {
		tx.Rollback()
		batch.Rollback()
		os.RemoveAll(path)
	}()

	g.RawEvmItems.ForEach(func(key, value []byte) bool {
		if err != nil {
			return false
		}

		err = batch.Put(kv.PlainState, key, value)
		if err != nil {
			return false
		}

		if batch.BatchSize() >= int(batchSize) {
			if err = batch.Commit(); err != nil {
				return false
			}
		}

		return true
	})
	if err != nil {
		return err
	}
	if err := batch.Commit(); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	if err := s.legacyEvmDBFromErigon(); err != nil {
		return fmt.Errorf("Unable to convert legacy erigon EVMDB to legacy one, err:%q", err)
	}

	return nil
}

// TODO find a correct place to close LegacyEvmDb properly
func (s *Store) legacyEvmDBFromErigon() error {

	legacyStateDB, err := gethstate.New(common.Hash{}, s.LegacyEvmState, nil)
	if err != nil {
		return fmt.Errorf("Unable to instantiate legacy StateDB")
	}

	roTx, err := s.EvmDb.RwKV().BeginRo(context.Background())
	if err != nil {
		return err
	}

	defer roTx.Rollback()

	// reads kv.Plainstate and pushes into legacyStateWriter
	state.DumpToCollector(&legacyStateWriter{legacyStateDB, 0}, nil, roTx)
	//panic("breakpoint")
	_, err = legacyStateDB.Commit(true)
	if err != nil {
		return err
	}

	//s.LegacyEvmDb = LegacyEvmDb
	//s.LegacyEvmState = LegacyEvmState

	return nil
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

type legacyStateWriter struct {
	dst  *gethstate.StateDB
	accs int
}

// OnRoot is called with the state root
func (w *legacyStateWriter) OnRoot(root common.Hash) {
}

// test it
// OnAccount is called once for each account in the trie
func (w *legacyStateWriter) OnAccount(addr common.Address, acc state.DumpAccount) {
	w.dst.SetNonce(addr, acc.Raw.Nonce)
	w.dst.SetBalance(addr, acc.Raw.Balance)
	if acc.Code != nil {
		w.dst.SetCode(addr, acc.Code)
	}
	for k, v := range acc.Storage {
		w.dst.SetState(addr, k, common.HexToHash(v))
	}
	w.accs++

	if w.accs%1000 == 0 {
		// flush data
		_ = w.dst.IntermediateRoot(false)
	}
}
