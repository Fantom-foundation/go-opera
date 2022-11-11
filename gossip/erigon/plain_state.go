package erigon

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/c2h5oh/datasize"

	"github.com/holiman/uint256"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/state/snapshot"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/trie"

	"github.com/Fantom-foundation/go-opera/gossip/evmstore"
	"github.com/Fantom-foundation/go-opera/logger"

	"github.com/Fantom-foundation/lachesis-base/inter/idx"

	elog "github.com/ledgerwatch/log/v3"

	ecommon "github.com/ledgerwatch/erigon/common"
	"github.com/ledgerwatch/erigon/common/dbutils"
	eaccounts "github.com/ledgerwatch/erigon/core/types/accounts"
	"github.com/ledgerwatch/erigon/migrations"
	"github.com/ledgerwatch/erigon/params"

	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/erigon-lib/kv/mdbx"
)

const bufferOptimalSize = 500 * datasize.MB

// find out config.DatabaseVerbosity, config.MdbxPageSize
func openDatabase(logger logger.Instance, label kv.Label) (kv.RwDB, error) {
	var name string
	switch label {
	case kv.ChainDB:
		name = "chaindata"
	case kv.TxPoolDB:
		name = "txpool"
	default:
		name = "test"
	}
	var db kv.RwDB

	dbPath := filepath.Join(DefaultDataDir(), "erigon", name)

	var openFunc func(exclusive bool) (kv.RwDB, error)
	logger.Log.Info("Opening Database", "label", name, "path", dbPath)
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

// MakeChainDatabase open a database using the flags passed to the client and will hard crash if it fails.
func MakeChainDatabase(logger logger.Instance) kv.RwDB {
	chainDb, err := openDatabase(logger, kv.TxPoolDB)
	if err != nil {
		Fatalf("Could not open database: %v", err)
	}
	return chainDb
}

// no duplicated keys are allowed
func ReadErigonTableNoDups(table string, tx kv.Tx) error {

	start := time.Now()
	logEvery := time.NewTicker(30 * time.Second)
	defer logEvery.Stop()

	c, err := tx.Cursor(table)
	if err != nil {
		return err
	}
	defer c.Close()

	records := 0
	for k, _, e := c.First(); k != nil; k, _, e = c.Next() {
		if e != nil {
			return e
		}

		records += 1
		//fmt.Printf("%x => %x\n", k, v)
	}
	log.Info("Reading table", table, "is complete", "elapsed", common.PrettyDuration(time.Since(start)), "records", records)
	return nil
}

func GeneratePlainState(root common.Hash, chaindb ethdb.KeyValueStore, db kv.RwDB, lastBlockIdx idx.Block) (err error) {
	return traverseSnapshot(chaindb, root, db)
}

func traverseSnapshot(diskdb ethdb.KeyValueStore, root common.Hash, db kv.RwDB) error {
	snaptree, err := snapshot.New(diskdb, trie.NewDatabase(diskdb), 256, root, false, false, false)
	if err != nil {
		return fmt.Errorf("Unable to build a snaptree, err: %q", err)
	}

	accIt, err := snaptree.AccountIterator(root, common.Hash{})
	if err != nil {
		return fmt.Errorf("Unable to make account iterator from snaptree, err: %q", err)
	}
	defer accIt.Release()

	log.Info("Snapshot traversal started", "root", root.Hex())
	var (
		start                    = time.Now()
		accounts                 uint64
		missingAddresses         int
		missingContractCode      int
		validContractAccounts    int
		validNonContractAccounts int
		invalidAccounts1         int
		invalidAccounts2         int
		storageRecords           int
		logEvery                 = time.NewTicker(60 * time.Second)
	)

	defer logEvery.Stop()

	// declare erigon buffer
	buf := newAppendBuffer(bufferOptimalSize)

	/* Plan
	   1. Loop over Account Iterator
	   2. Map account hash into address by invoking addressFromPreimage
	   3. Put account data or storage into erigon buffer
	   4. Sort items in buffer for efficient writing data into erigon table
	   5. Iterate over buf items and write them into erigon PlainState
	*/
	for accIt.Next() {
		storageIterations := 0
		accHash := accIt.Hash()

		addr, err := addressFromPreimage(db, ecommon.Hash(accHash))
		if err != nil {
			return fmt.Errorf("unable to get address from preimage, err: %q", err)
		}

		snapAccount, err := snapshot.FullAccount(accIt.Account())
		if err != nil {
			return fmt.Errorf("Unable to get snapshot.Account from account Iterator, err: %q", err)
		}

		stateAccount := state.Account{
			Nonce:    snapAccount.Nonce,
			Balance:  snapAccount.Balance,
			Root:     common.BytesToHash(snapAccount.Root),
			CodeHash: snapAccount.CodeHash,
		}

		switch {
		case stateAccount.Root != types.EmptyRootHash && !bytes.Equal(stateAccount.CodeHash, evmstore.EmptyCode):
			// contract account
			validContractAccounts++
			eAccount := transformStateAccount(stateAccount, true)

			// writing data and storage into buf
			storageIterations, err = putAccountDataStorageToBuf(buf, storageRecords, eAccount, snaptree, addr, root, accHash)
			if err != nil {
				return err
			}

		case stateAccount.Root == types.EmptyRootHash && bytes.Equal(stateAccount.CodeHash, evmstore.EmptyCode):
			// non contract account
			validNonContractAccounts++
			eAccount := transformStateAccount(stateAccount, false)
			if err := putAccountDataToBuf(buf, eAccount, addr); err != nil {
				return err
			}
		case stateAccount.Root != types.EmptyRootHash && bytes.Equal(stateAccount.CodeHash, evmstore.EmptyCode):
			// root of storage trie is not empty , but codehash is empty
			// looks like it is invalid account
			// invalidAccounts1  = 0 forget about this case
			invalidAccounts1++
			code := rawdb.ReadCode(diskdb, common.BytesToHash(stateAccount.CodeHash))
			if len(code) == 0 {
				missingContractCode++
				//log.Error("Code is missing", "hash", common.BytesToHash(stateAccount.CodeHash))
				//return errors.New("missing code")
			}

			eAccount := transformStateAccount(stateAccount, true)

			storageIterations, err = putAccountDataStorageToBuf(buf, storageRecords, eAccount, snaptree, addr, root, accHash)
			if err != nil {
				return err
			}

		case stateAccount.Root == types.EmptyRootHash && !bytes.Equal(stateAccount.CodeHash, evmstore.EmptyCode):
			// invalid accounts2=407
			// TODO address it https://blog.ethereum.org/2020/07/17/ask-about-geth-snapshot-acceleration/
			// Self-destructs (and deletions) are special beasts as they need to short circuit diff layer descent.
			invalidAccounts2++
			eAccount := transformStateAccount(stateAccount, true)

			// writing data and storage
			storageIterations, err = putAccountDataStorageToBuf(buf, storageRecords, eAccount, snaptree, addr, root, accHash)
			if err != nil {
				return err
			}

		}
		storageRecords += storageIterations
		accounts++
	}

	if missingAddresses > 0 {
		log.Warn("Snapshot traversal is incomplete due to missing addresses", "missing", missingAddresses)
	}

	log.Info("Snapshot traversal is complete", "accounts", accounts,
		"elapsed", common.PrettyDuration(time.Since(start)), "missingContractCode", missingContractCode)

	log.Info("Valid", "Contract accounts: ", validContractAccounts, "Storage records", storageRecords, "Valid non contract accounts", validNonContractAccounts,
		"invalid accounts1", invalidAccounts1, "invalid accounts2", invalidAccounts2)

	// Sort data in buffer
	log.Info("Buffer length", "is", buf.Len())
	log.Info("Sorting data in buffer started...")
	start = time.Now()
	// there are no duplicates keys in buf
	buf.Sort()
	log.Info("Sorting data is complete", "elapsed", common.PrettyDuration(time.Since(start)))

	start = time.Now()
	if err := buf.writeIntoTable(db, kv.PlainState); err != nil {
		return err
	}

	log.Info("Writing data into erigon PlainState is complete", "elapsed", common.PrettyDuration(time.Since(start)))
	return nil
}

func putAccountDataToBuf(buf *appendSortableBuffer, acc eaccounts.Account, addr ecommon.Address) error {
	key := addr.Bytes()
	value := make([]byte, acc.EncodingLengthForStorage())
	acc.EncodeForStorage(value)
	return buf.Put(key, value)
}

func putAccountStorageToBuf(buf *appendSortableBuffer, incarnation uint64, addr ecommon.Address, key *ecommon.Hash, val *uint256.Int) error {
	compositeKey := dbutils.PlainGenerateCompositeStorageKey(addr.Bytes(), incarnation, key.Bytes())
	value := val.Bytes()
	return buf.Put(compositeKey, value)

}

// putAccountDataStorageToBuf puts account data into buf. Additionally, it runs sotrage itarator and writes storage into buf.
func putAccountDataStorageToBuf(buf *appendSortableBuffer, storageRecords int, eAccount eaccounts.Account, snapTree *snapshot.Tree, addr ecommon.Address, root, accHash common.Hash) (int, error) {
	iterations := 0
	if err := putAccountDataToBuf(buf, eAccount, addr); err != nil {
		return iterations, err
	}

	stIt, err := snapTree.StorageIterator(root, accHash, common.Hash{})
	if err != nil {
		return iterations, err
	}

	defer stIt.Release()

	for stIt.Next() {
		// to make sure it is a right way to write storage
		key, value := ecommon.Hash(stIt.Hash()), uint256.NewInt(0).SetBytes(stIt.Slot())
		if err := putAccountStorageToBuf(buf, eAccount.Incarnation, addr, &key, value); err != nil {
			return iterations, err
		}
		iterations += 1
	}

	return iterations, nil
}

// transformStateAccount transforms state.Account into erigon Account representation
func transformStateAccount(account state.Account, isContractAcc bool) (eAccount eaccounts.Account) {
	eAccount.Initialised = true // ?
	bal, overflow := uint256.FromBig(account.Balance)
	if overflow {
		panic(fmt.Sprintf("overflow occured while converting from account.Balance"))
	}
	eAccount.Nonce = account.Nonce
	eAccount.Balance = *bal
	eAccount.Root = ecommon.Hash(account.Root)
	eAccount.CodeHash = ecommon.Hash(common.BytesToHash(account.CodeHash))
	if isContractAcc {
		eAccount.Incarnation = 1
	}

	return
}
