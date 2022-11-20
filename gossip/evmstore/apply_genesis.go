package evmstore

import (
	"context"
	"os"
	"path/filepath"

	"github.com/Fantom-foundation/go-opera/gossip/evmstore/state"
	"github.com/Fantom-foundation/go-opera/opera/genesis"

	"github.com/Fantom-foundation/go-opera/erigon"
	"github.com/Fantom-foundation/go-opera/gossip/evmstore/ethdb"
	"github.com/ethereum/go-ethereum/common"

	"github.com/ledgerwatch/erigon-lib/kv"

	"github.com/c2h5oh/datasize"

	gethethdb "github.com/ethereum/go-ethereum/ethdb"
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

	roTx, err := s.EvmDb.RwKV().BeginRo(context.Background())
	if err != nil {
		return err
	}

	defer roTx.Rollback()

	// legacy StateDB
	//
	dst, err := s.StateDB(hash.Hash{})
	if err != nil {
		return
	}

	state.DumpToCollector(&stateWriter{dst, 0}, nil, roTx)
	root, err := dst.Commit(true)
	if err != nil {
		return
	}

	return nil
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

type stateWriter struct {
	dst  *state.StateDB // think of makeing evmcore.StateDB
	accs int
}

// OnRoot is called with the state root
func (w *stateWriter) OnRoot(root common.Hash) {
}

// OnAccount is called once for each account in the trie
func (w *stateWriter) OnAccount(addr common.Address, acc state.DumpAccount) {
	w.dst.SetNonce(addr, acc.Raw.Nonce)
	w.dst.SetBalance(addr, acc.Raw.Balance)
	if acc.Code != nil {
		w.dst.SetCode(addr, acc.Code)
	}
	for k, v := range acc.Storage {
		w.dst.SetState(addr, k, common.HexToHash(v))
	}
	w.accs++
}
