package erigon

import (
	"context"
	"io"
	"path/filepath"

	"github.com/c2h5oh/datasize"

	"github.com/Fantom-foundation/go-opera/logger"
	"github.com/Fantom-foundation/lachesis-base/common/bigendian"
	elog "github.com/ledgerwatch/log/v3"

	"github.com/ledgerwatch/erigon/migrations"
	"github.com/ledgerwatch/erigon/params"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/erigon-lib/kv/mdbx"
)

var emptyCode = crypto.Keccak256(nil)

// openDatabase opens lmdb database using specified label
func openDatabase(logger logger.Instance, label kv.Label) (kv.RwDB, error) {
	var name string
	switch label {
	case kv.ChainDB:
		name = "chaindata"
	case kv.TxPoolDB:
		name = "txpool"
	case kv.ConsensusDB: //fakenet
		name = "consensusDB"
	default:
		name = "test"
	}
	var db kv.RwDB

	dbPath := filepath.Join(DefaultDataDir(), "erigon", name)

	var openFunc func(exclusive bool) (kv.RwDB, error)
	logger.Log.Info("Opening Erigon Database", "label", name, "path", dbPath)
	elog := elog.New()
	openFunc = func(exclusive bool) (kv.RwDB, error) {
		opts := mdbx.NewMDBX(elog).Path(dbPath).Label(label).DBVerbosity( /*config.DatabaseVerbosity*/ 0).MapSize(6 * datasize.TB)
		if exclusive {
			opts = opts.Exclusive()
		}
		if label == kv.ChainDB {
			opts = opts.PageSize( /*config.MdbxPageSize.Bytes()*/ 100000000000)
		}
		return opts.Open()
	}
	var err error
	db, err = openFunc(false)
	if err != nil {
		return nil, err
	}
	migrator := migrations.NewMigrator(label)
	if err := migrator.VerifyVersion(db); err != nil {
		return nil, err
	}

	has, err := migrator.HasPendingMigrations(db)
	if err != nil {
		return nil, err
	}
	if has {
		elog.Info("Re-Opening DB in exclusive mode to apply migrations")
		db.Close()
		db, err = openFunc(true)
		if err != nil {
			return nil, err
		}
		if err = migrator.Apply(db, DefaultDataDir()); err != nil {
			return nil, err
		}
		db.Close()
		db, err = openFunc(false)
		if err != nil {
			return nil, err
		}
	}

	if err := db.Update(context.Background(), func(tx kv.RwTx) (err error) {
		return params.SetErigonVersion(tx, params.VersionKeyCreated)
	}); err != nil {
		return nil, err
	}

	return db, nil
}

// MakeChainDatabase opens a database and it crashes if it fails to open
func MakeChainDatabase(logger logger.Instance, label kv.Label) kv.RwDB {
	chainDb, err := openDatabase(logger, label)
	if err != nil {
		utils.Fatalf("Could not open database: %v", err)
	}
	return chainDb
}

// Write iterates over erigon kv.PlainState records and populates io.Writer
func Write(writer io.Writer, tx kv.Tx) (accounts int, err error) {
	c, err := tx.Cursor(kv.PlainState)
	if err != nil {
		return accounts, err
	}
	defer c.Close()

	for k, v, e := c.First(); k != nil; k, v, e = c.Next() {
		if e != nil {
			return accounts, e
		}

		_, err := writer.Write(bigendian.Uint32ToBytes(uint32(len(k))))
		if err != nil {
			return accounts, err
		}
		_, err = writer.Write(k)
		if err != nil {
			return accounts, err
		}
		_, err = writer.Write(bigendian.Uint32ToBytes(uint32(len(v))))
		if err != nil {
			return accounts, err
		}
		_, err = writer.Write(v)
		if err != nil {
			return accounts, err
		}
		accounts++
	}

	return accounts, nil
}
