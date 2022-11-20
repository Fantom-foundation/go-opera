package erigon

import (
	"context"
	"path/filepath"
	"strconv"

	"github.com/ethereum/go-ethereum/cmd/utils"

	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/erigon-lib/kv/mdbx"

	"github.com/ledgerwatch/erigon/migrations"
	"github.com/ledgerwatch/erigon/params"

	"github.com/Fantom-foundation/go-opera/logger"

	"github.com/c2h5oh/datasize"

	elog "github.com/ledgerwatch/log/v3"

	"github.com/ledgerwatch/log/v3"
)

// openDatabase opens lmdb database using specified label
func openDatabase(logger logger.Instance, label kv.Label, inMem bool, erigonDbId uint) (db kv.RwDB, err error) {
	var (
		name string
	)

	switch label {
	case kv.ChainDB: //chainKV
		name = "chaindata"
	case kv.TxPoolDB:
		name = "txpool" // tempDB
	case kv.ConsensusDB: //genesisKV
		name = "consensusDB"
	default:
		name = "test"
	}

	switch {
	case erigonDbId > 0:
		id := strconv.Itoa(int(erigonDbId))
		name += id
	default:
	}

	dbPath := filepath.Join(DefaultDataDir(), "erigon", name)

	var openFunc func(exclusive bool) (kv.RwDB, error)
	logger.Log.Info("Opening Erigon Database", "label", name, "path", dbPath)
	elog := elog.New()
	openFunc = func(exclusive bool) (kv.RwDB, error) {
		opts := mdbx.NewMDBX(elog).Path(dbPath).Label(label).DBVerbosity( /*config.DatabaseVerbosity*/ 0).MapSize(6 * datasize.TB)
		if exclusive {
			opts = opts.Exclusive()
		}
		if inMem {
			opts = opts.InMem()
		}
		if label == kv.ChainDB {
			opts = opts.PageSize( /*config.MdbxPageSize.Bytes()*/ 100000000000)
		}
		return opts.Open()
	}

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

	return
}

// MakeChainDatabase opens a database and it crashes if it fails to open
func MakeChainDatabase(logger logger.Instance, label kv.Label, inMem bool, erigonDbId uint) kv.RwDB {
	chainDb, err := openDatabase(logger, label, inMem, erigonDbId)
	if err != nil {
		utils.Fatalf("Could not open database: %v", err)
	}
	return chainDb
}

func OpenChainDatabase(path string, logger log.Logger, inMem bool, readonly bool) kv.RwDB {
	opts := mdbx.NewMDBX(logger).Label(kv.ChainDB)
	if readonly {
		opts = opts.Readonly()
	}
	if inMem {
		opts = opts.InMem()
	} else {
		opts = opts.Path(path)
	}

	return opts.MustOpen()
}
