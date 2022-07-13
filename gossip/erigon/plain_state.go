package erigon

import (
	"path/filepath"
	"time"
	"bytes"
	"errors"
	"fmt"
	"context"

	"github.com/c2h5oh/datasize"

	"github.com/holiman/uint256"


	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/state/snapshot"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	//"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"


	"github.com/Fantom-foundation/go-opera/gossip/evmstore"
	"github.com/Fantom-foundation/go-opera/logger"

	"github.com/Fantom-foundation/lachesis-base/inter/idx"

	elog "github.com/ledgerwatch/log/v3"

	ecommon "github.com/ledgerwatch/erigon/common"
	"github.com/ledgerwatch/erigon/common/dbutils"
	eaccounts "github.com/ledgerwatch/erigon/core/types/accounts"
	"github.com/ledgerwatch/erigon/params"
	"github.com/ledgerwatch/erigon/migrations"
	//"github.com/ledgerwatch/erigon/crypto"

	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/erigon-lib/kv/mdbx"

)

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


	dbPath := filepath.Join(defaultDataDir(), "erigon", name)
	
	var openFunc func(exclusive bool) (kv.RwDB, error)
	logger.Log.Info("Opening Database", "label", name, "path", dbPath)
	elog := elog.New()
	openFunc = func(exclusive bool) (kv.RwDB, error) {
		opts := mdbx.NewMDBX(elog).Path(dbPath).Label(label).DBVerbosity(/*config.DatabaseVerbosity*/ 0).MapSize(6 * datasize.TB)
		if exclusive {
			opts = opts.Exclusive()
		}
		if label == kv.ChainDB {
			opts = opts.PageSize(/*config.MdbxPageSize.Bytes()*/ 100000000000)
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
		if err = migrator.Apply(db, defaultDataDir()); err != nil {
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
	chainDb, err := openDatabase(logger, kv.ChainDB)
	if err != nil {
		Fatalf("Could not open database: %v", err)
	}
	return chainDb
}

// CursorDupSort only, TODO to make it for usual cursor as well
func ReadErigonTable(table string, tx kv.Tx) error {

	start := time.Now()
	logEvery := time.NewTicker(30 * time.Second)
	defer logEvery.Stop()


	c, err := tx.CursorDupSort(table)
	if err != nil {
		return err
	}
	defer c.Close()

	
	dupRecords, records := 0, 0
	for k, _, e := c.First(); k != nil; k, _, e = c.NextNoDup() {
		if e != nil {
			return e
		}

		records += 1
		// run for loop for duplicates
		for vds, eds :=  c.FirstDup(); vds != nil ; _, vds, eds = c.NextDup() {
			if eds != nil {
				return eds
			}
			dupRecords += 1
		} 

		//fmt.Printf("%x => %x\n", k, v)
	}
	log.Info("Reading", table,  "is complete", "elapsed", common.PrettyDuration(time.Since(start)), "records", records, "dupRecords", dupRecords)
	return nil
}


/*
func SetupDB() (kv.RwDB, string, error) {
	tmpDir := filepath.Join(os.TempDir(), "lmdb")
	db, err := mdbx.NewMDBX(elog.New()).
			Path(tmpDir).
			WithTablessCfg(func(defaultBuckets kv.TableCfg) kv.TableCfg {
				return kv.TableCfg{
					kv.PlainState: kv.TableCfgItem{},
					kv.TrieOfAccounts: kv.TableCfgItem{},
					kv.TrieOfStorage: kv.TableCfgItem{},
					kv.HashedStorage: kv.TableCfgItem{},
					kv.HashedAccounts: kv.TableCfgItem{},
		        }
			}).Open()

	return db, tmpDir, err
}
*/

func GeneratePlainState(mptFlag string, root common.Hash, chaindb ethdb.KeyValueStore, db kv.RwDB, lastBlockIdx idx.Block ) (err error) {
	switch mptFlag {
	case "mpt":
		//err = traverseMPT(chaindb, root, db, lastBlockIdx)
	case "snap":
		err = traverseSnapshot(chaindb, root, db)
	default:
		err = errors.New("--mpt.traversal.mode must be one of {mpt, snap}")
	}
	return 
}

// Attention! This function does not work properly.
/*
func traverseMPT(diskdb ethdb.KeyValueStore, root common.Hash, db kv.RwDB, lastBlockIdx idx.Block) error {
	triedb := trie.NewDatabase(diskdb)
	t, err := trie.NewSecure(root, triedb)
	if err != nil {
		log.Error("Failed to open trie", "root", root, "err", err)
		return err
	}

	log.Info("Start traversing the state", "root", root, "number", lastBlockIdx)
	var (
		missingPreimages int
		accounts   int
		slots      int
		codes      int
		lastReport time.Time
		start      = time.Now()
	)
	accIter := trie.NewIterator(t.NodeIterator(nil))

	for accIter.Next() {
		accounts += 1
		var stateAccount state.Account

		if err := rlp.DecodeBytes(accIter.Value, &stateAccount); err != nil {
			log.Error("Invalid account encountered during traversal", "err", err)
			return err
		}

		
		/*
		addrBytes := t.GetKey(accIter.Key)
		if addrBytes == nil {
			missingPreimages++
			log.Info("preimage is missing")
		}
		*/
		/*
		addr := ecommon.BytesToAddress(accIter.Key)
		log.Info("Addr", addr.Hex())
		if len(addr) != 20 {
			log.Warn("address is invalid")
		}

		switch {
		case stateAccount.Root != types.EmptyRootHash && !bytes.Equal(stateAccount.CodeHash, evmstore.EmptyCode):
			// contract accoount
			log.Info("contract account")
			codes += 1
			eAccount := transformStateAccount(stateAccount, true)

			if err := writeAccountData(db, eAccount, addr); err != nil {
				return err
			}

            // make sure it is a correct iterator
			storageIt := trie.NewIterator(t.NodeIterator(nil))
			for storageIt.Next() {
				slots += 1
				_, content, _, err := rlp.Split(storageIt.Value)
				if err != nil {
					log.Error("Failed to decode the value returned by iterator", "error", err)
					continue
				}

				// to make sure it is a right way to write storage
				key, value := ecommon.BytesToHash(t.GetKey(storageIt.Key)), uint256.NewInt(0).SetBytes(content)
				if err := writeAccountStorage(db, eAccount.Incarnation, addr, &key, value); err != nil {
					return err
				}
			}

		case stateAccount.Root == types.EmptyRootHash && bytes.Equal(stateAccount.CodeHash, evmstore.EmptyCode):
			// non contract account
			log.Info("non contract account")
			eAccount := transformStateAccount(stateAccount, false)
			if err := writeAccountData(db, eAccount, addr); err != nil {
				return err
			}

		default:
			// TODO address this case
			log.Info("default case")
			continue
			
		}


		/*
		if acc.Root != types.EmptyRootHash {
			storageTrie, err := trie.NewSecure(acc.Root, triedb)
			if err != nil {
				log.Error("Failed to open storage trie", "root", acc.Root, "err", err)
				return err
			}
			storageIter := trie.NewIterator(storageTrie.NodeIterator(nil))
			for storageIter.Next() {
				slots += 1
			}
			if storageIter.Err != nil {
				log.Error("Failed to traverse storage trie", "root", acc.Root, "err", storageIter.Err)
				return storageIter.Err
			}
		}
		// contract account
		if !bytes.Equal(acc.CodeHash, evmstore.EmptyCode) {
			code := rawdb.ReadCode(chaindb, common.BytesToHash(acc.CodeHash))
			if len(code) == 0 {
				log.Error("Code is missing", "hash", common.BytesToHash(acc.CodeHash))
				return errors.New("missing code")
			}
		
	
			codes += 1
		} 
		*/
/*
		if time.Since(lastReport) > time.Second*8 {
			log.Info("Traversing MPT", "accounts", accounts, "slots", slots, "codes", codes, "elapsed", common.PrettyDuration(time.Since(start)))
			lastReport = time.Now()
		}
	}

	if accIter.Err != nil {
		log.Error("Failed to traverse state trie", "root", root, "err", accIter.Err)
		return accIter.Err
	}

	if missingPreimages > 0 {
		log.Warn("MPT traversal is incomplete due to missing preimages", "missing", missingPreimages)
	}

	log.Info("MPT traversal is complete", "accounts", accounts, "slots", slots, "codes", codes, "elapsed", common.PrettyDuration(time.Since(start)))



	return nil
}
*/


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
		start    = time.Now()
		accounts uint64
		missingAddresses int
		missingContractCode int
		validContractAccounts int
		validNonContractAccounts int 
		invalidAccounts1 int
		invalidAccounts2 int
		matchedAccounts, notMatchedAccounts uint64
		logEvery = time.NewTicker(60 * time.Second)
		bufferOptimalSize = 500 * datasize.MB

		
	)
	defer logEvery.Stop()


	buf := newAppendBuffer(bufferOptimalSize)

	for accIt.Next() {
		accHash := accIt.Hash()

	
		addr, err := addressFromPreimage(db, accHash)
		if err != nil {
			return fmt.Errorf("unable to get address from preimage, err: %q", err)
		}
	
		snapAccount, err := snapshot.FullAccount(accIt.Account())
		if err != nil {
			return fmt.Errorf("Unable to get snapshot.Account from account Iterator, err: %q", err)
		}

		stateAccount := state.Account{
			Nonce: snapAccount.Nonce,
			Balance: snapAccount.Balance,
			Root: common.BytesToHash(snapAccount.Root),
			CodeHash: snapAccount.CodeHash,
		}
			
		
		switch {
			case stateAccount.Root != types.EmptyRootHash && !bytes.Equal(stateAccount.CodeHash, evmstore.EmptyCode):
				//log.Info("contract account is valid")
				validContractAccounts++
				eAccount := transformStateAccount(stateAccount, true)

				// writing data and storage
				if err := putAccountDataStorageToBuf(buf, eAccount, snaptree, addr, root, accHash); err != nil {
					return err
				}

			
			case stateAccount.Root == types.EmptyRootHash && bytes.Equal(stateAccount.CodeHash, evmstore.EmptyCode):
				// non contract account
				//log.Info("non contract account is valid")
				validNonContractAccounts++
				eAccount := transformStateAccount(stateAccount, false)
				putAccountDataToBuf(buf, eAccount, addr)
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

				putAccountDataStorageToBuf(buf, eAccount, snaptree, addr, root, accHash)

			
			case stateAccount.Root == types.EmptyRootHash && !bytes.Equal(stateAccount.CodeHash, evmstore.EmptyCode):
				// invalid accounts2=407
				// TODO address it https://blog.ethereum.org/2020/07/17/ask-about-geth-snapshot-acceleration/
				// Self-destructs (and deletions) are special beasts as they need to short circuit diff layer descent.
				invalidAccounts2++
				eAccount := transformStateAccount(stateAccount, true)

				// writing data and storage
				putAccountDataStorageToBuf(buf, eAccount, snaptree, addr, root, accHash)

		}
		accounts++
			
		/*
		if checkAcc && accounts == uint64(accountLimit) {
			log.Info("Break", "Accounts", accounts, "accountLimit", accountLimit)
			break
		}
		*/

		/*
		select {
		default:
		case <-logEvery.C:
			log.Info("Snapshot traversing in progress", "at", accIt.Hash(), "accounts", 
			accounts, 
			"Preimages matched Accounts", matchedAccounts, "Not Matched Accounts", notMatchedAccounts,
				"elapsed", common.PrettyDuration(time.Since(start)))
		}
		*/
	}
	

	
	if missingAddresses > 0 {
		log.Warn("Snapshot traversal is incomplete due to missing addresses", "missing", missingAddresses)
	}

	log.Info("Snapshot traversal is complete", "accounts", accounts,
		"elapsed", common.PrettyDuration(time.Since(start)), "missingContractCode", missingContractCode)

	log.Info("Preimages matching is complete", "Matched Accounts", matchedAccounts, "Not Matched Accounts", notMatchedAccounts)

	log.Info("Valid", "Contract accounts: ", validContractAccounts, "Valid non contract accounts", validNonContractAccounts,
	 "invalid accounts1", invalidAccounts1, "invalid accounts2" , invalidAccounts2)

	log.Info("Buf size", "is", buf.size)
	log.Info("Sorting data in buffer")
	start = time.Now()
	buf.Sort()
	log.Info("Sorting data is complete", "elapsed", common.PrettyDuration(time.Since(start)))

	
	tx, err := db.BeginRw(context.Background())
	if err != nil {
		return err
	}

	defer tx.Rollback() 

	c, err := tx.RwCursor(kv.PlainState)
	if err != nil {
		return err 
	}

	defer c.Close()

	log.Info("Iterate over sorted entries and write them into kv.Plainstate")
	start = time.Now()
	for _, entry := range buf.sortedBuf {
		if err := c.Append(entry.key, entry.value); err != nil {
			return err
		}
	}
	log.Info("Writing data is complete", "elapsed", common.PrettyDuration(time.Since(start)))

	
	return tx.Commit()
}


func putAccountDataToBuf(buf *appendSortableBuffer, acc eaccounts.Account, addr ecommon.Address) {
	/*
	return db.Update(context.Background(), func(tx kv.RwTx) error {
		value := make([]byte, acc.EncodingLengthForStorage())
		acc.EncodeForStorage(value)
		return tx.Put(kv.PlainState, addr[:], value)
	})
	*/
	key := addr.Bytes()
	value := make([]byte, acc.EncodingLengthForStorage())
	acc.EncodeForStorage(value)
	buf.Put(key, value)
}

// ask about how to write in more efficient way using RwCursor 
func putAccountStorageToBuf(buf *appendSortableBuffer, incarnation uint64, addr ecommon.Address, key *ecommon.Hash, val *uint256.Int) {
	/*
	return db.Update(context.Background(), func(tx kv.RwTx) error {
		compositeKey := dbutils.PlainGenerateCompositeStorageKey(addr.Bytes(), incarnation, key.Bytes())
		value := val.Bytes()
		return tx.Put(kv.PlainState, compositeKey, value)
	})
	*/
	compositeKey := dbutils.PlainGenerateCompositeStorageKey(addr.Bytes(), incarnation, key.Bytes())
	value := val.Bytes()
	buf.Put(compositeKey,value)

}

func putAccountDataStorageToBuf(buf *appendSortableBuffer, eAccount eaccounts.Account, snapTree *snapshot.Tree, addr ecommon.Address, root, accHash common.Hash)  error {
	
	putAccountDataToBuf(buf, eAccount, addr)

	stIt, err := snapTree.StorageIterator(root, accHash, common.Hash{})
	if err != nil {
		return err
	}

	defer stIt.Release()

	for stIt.Next() {
		// to make sure it is a right way to write storage
		key, value := ecommon.Hash(stIt.Hash()), uint256.NewInt(0).SetBytes(stIt.Slot())
		putAccountStorageToBuf(buf, eAccount.Incarnation, addr, &key, value)
	}

	return nil
}



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