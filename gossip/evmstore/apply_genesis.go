package evmstore

import (
	"context"
	"math/rand"
	"os"
	"path/filepath"

	//"github.com/Fantom-foundation/lachesis-base/kvdb"

	//"github.com/ledgerwatch/erigon/ethdb"

	"github.com/Fantom-foundation/go-opera/gossip/evmstore/state"
	"github.com/Fantom-foundation/go-opera/logger"
	"github.com/Fantom-foundation/go-opera/opera/genesis"

	"github.com/Fantom-foundation/go-opera/erigon"
	"github.com/Fantom-foundation/go-opera/gossip/evmstore/ethdb"
	"github.com/ethereum/go-ethereum/common"

	"github.com/ledgerwatch/erigon-lib/kv"
	//"github.com/ledgerwatch/erigon/ethdb/olddb"

	estate "github.com/ledgerwatch/erigon/core/state"
	erigonethdb "github.com/ledgerwatch/erigon/ethdb"

	"github.com/c2h5oh/datasize"
)

const batchSizeStr = "256M"

// ApplyGenesis writes initial state.
func (s *Store) ApplyGenesis(g genesis.Genesis) (err error) {

	r := rand.Intn(100)

	tempDB := erigon.MakeChainDatabase(logger.New("tempDB"), kv.TxPoolDB, uint(r))
	defer tempDB.Close()

	tx, err := tempDB.BeginRw(context.Background())
	if err != nil {
		return err
	}

	var batchSize datasize.ByteSize
	must(batchSize.UnmarshalText([]byte(batchSizeStr)))

	//var batch ethdb.DbWithPendingMutations
	// state is stored through ethdb batches
	path := filepath.Join(erigon.DefaultDataDir(), "erigon", "batch")
	//batch = olddb.NewHashBatch(tx, nil, path, nil, nil)
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

	/* TODO make it match with Genesis stateroot
	root, err := erigon.CalcRoot("ApplyGenesis, stateRoot", tx)
	if err != nil {
		return err
	}
	log.Info("ApplyGenesis", "stateRoot", root.Hex())
	*/

	src := state.NewWithDatabase(erigonethdb.Database(batch))
	dst := state.NewWithStateReader(estate.NewPlainStateReader(tx))

	src.DumpToCollector(&stateWriter{dst, 0}, nil, tx)

	return nil
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

func must(err error) {
	if err != nil {
		panic(err)
	}
}
