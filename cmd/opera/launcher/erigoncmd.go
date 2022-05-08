package launcher

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path"
	"path/filepath"
	"time"

	"fmt"

	"github.com/holiman/uint256"
	"gopkg.in/urfave/cli.v1"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/state/snapshot"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/ethereum/go-ethereum/ethdb"

	"github.com/Fantom-foundation/go-opera/gossip/erigon"

	"github.com/Fantom-foundation/go-opera/gossip/evmstore"
	"github.com/Fantom-foundation/go-opera/integration"

	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/erigon-lib/kv/mdbx"

	ecommon "github.com/ledgerwatch/erigon/common"
	"github.com/ledgerwatch/erigon/common/dbutils"
	eaccounts "github.com/ledgerwatch/erigon/core/types/accounts"
	"github.com/ledgerwatch/erigon/crypto"
	erigonDB "github.com/ledgerwatch/erigon/ethdb"
	elog "github.com/ledgerwatch/log/v3"
)

func writeEVMToErigon(ctx *cli.Context) error {

	// initiate erigon lmdb
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
	if err != nil {
		return err
	}
	defer db.Close()

	cfg := makeAllConfigs(ctx)

	rawProducer := integration.DBProducer(path.Join(cfg.Node.DataDir, "chaindata"), cacheScaler(ctx))
	gdb, err := makeRawGossipStore(rawProducer, cfg)
	if err != nil {
		log.Crit("DB opening error", "datadir", cfg.Node.DataDir, "err", err)
	}
	if gdb.GetHighestLamport() != 0 {
		log.Warn("Attempting genesis export not in a beginning of an epoch. Genesis file output may contain excessive data.")
	}
	defer gdb.Close()

	chaindb := gdb.EvmStore().EvmDb
	root := common.Hash(gdb.GetBlockState().FinalizedStateRoot)


	switch ctx.String(mptTraversalMode.Name) {
	case "mpt":
		return traverseMPT()
	case "snap":
		return traverseSnap(chaindb, root)
	default:
		return errors.New("--mpt.traversal.mode must be one of {mpt, snap}")

	}
	

	
	

	triedb := trie.NewDatabase(chaindb)


	
	t, err := trie.NewSecure(root, triedb)
	if err != nil {
		log.Error("Failed to open trie", "root", root, "err", err)
		return err
	}

	log.Info("Start traversing the state", "root", root, "number", gdb.GetBlockState().LastBlock.Idx)
	var (
		accounts   int
		slots      int
		codes      int
		lastReport time.Time
		start      = time.Now()
	)
	accIter := trie.NewIterator(t.NodeIterator(nil))

	

	

	/*
	randomAddr := func() ecommon.Address {
		key, _ := crypto.GenerateKey()
		addr := ecommon.Address(crypto.PubkeyToAddress(key.PublicKey))
		return addr
	}
	*/

   // generate PlainState
	for accIter.Next() {
		accounts += 1
		var acc state.Account

		if err := rlp.DecodeBytes(accIter.Value, &acc); err != nil {
			log.Error("Invalid account encountered during traversal", "err", err)
			return err
		}

		eAccount := transformStateAccount(acc)
		// TODO replace random accs 
		//addr := randomAddr()

		switch {
		// EOA non contract accounts
		case acc.Root == types.EmptyRootHash && bytes.Equal(acc.CodeHash, evmstore.EmptyCode):
			// write non-contract accounts to kv.PlainState
			if err := writeAccountData(db, eAccount, addr); err != nil {
				return err
			}
        
		// contract account
		case acc.Root != types.EmptyRootHash && !bytes.Equal(acc.CodeHash, evmstore.EmptyCode):
			// write contract accounts to kv.PlainState
			if err := writeAccountStorage(db, eAccount, addr); err != nil {
				return err
			}

		// TODO ADDRESS OTHER CASES If they are
		}

	
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

		if time.Since(lastReport) > time.Second*8 {
			log.Info("Traversing state", "accounts", accounts, "slots", slots, "codes", codes, "elapsed", common.PrettyDuration(time.Since(start)))
			lastReport = time.Now()
		}
	}

	if accIter.Err != nil {
		log.Error("Failed to traverse state trie", "root", root, "err", accIter.Err)
		return accIter.Err
	}

	log.Info("State is complete", "accounts", accounts, "slots", slots, "codes", codes, "elapsed", common.PrettyDuration(time.Since(start)))
	
	log.Info("Generate Hash State...")
	if err := erigon.GenerateHashedState("HashedState", db, tmpDir, context.Background()); err != nil {
		log.Error("GenerateHashedState error: ", err)
		return err
	}

	/*
	if err := generateHashState2(db); err != nil {
		// insert log.Error
		return err
	}
	*/
	// TODO insert timer
	log.Info("Generate Hash State is complete")
	log.Info("Calculating State Root...")
	trieCfg := erigon.StageTrieCfg(db, true, true, "", nil)
	hash, err := erigon.GenerateStateRoot("Intermediate Hashes", db, trieCfg)
	if err != nil {
		log.Error("GenerateIntermediateHashes error: ", err)
		return err
	}

	log.Info(fmt.Sprintf("[%s] Trie root", "GenerateStateRoot"), "hash", hash.Hex())
	/*
	root, err = CalcTrieRoot2(db)
	if err  != nil {
		log.Error("Failed to calculate state root", "root", root, "err", accIter.Err)
	}
	*/

	log.Info("Calculation of State Root Complete")


	return nil
}

func traverseMPT() error {

	return nil
}


func traverseSnapshot(diskdb ethdb.KeyValueStore, root common.Hash) error {
	snaptree, err := snapshot.New(diskdb, trie.NewDatabase(diskdb), 256, root, false, false, false)
	if err != nil {
		return err
	}

	accIt, err := snaptree.AccountIterator(root, common.Hash{})
	if err != nil {
		return err
	}
	defer accIt.Release()

	log.Info("Snapshot dumping started", "root", root)
	var (
		start    = time.Now()
		logged   = time.Now()
		accounts uint64
	)

	for accIt.Next() {
		account, err := snapshot.FullAccount(accIt.Account())
		if err != nil {
			return err
		}

		da := &state.DumpAccount{
			Balance:   account.Balance.String(),
			Nonce:     account.Nonce,
			Root:      account.Root,
			CodeHash:  account.CodeHash,
			SecureKey: accIt.Hash().Bytes(),
		}

		if len(da.SecureKey) == 0 {
			log.Info("Unable to find account address")
		} 

		daRootHash := common.BytesToHash(da.Root)
		// EOA non contract account
		if daRootHash == types.EmptyRootHash  && bytes.Equal(da.CodeHash, evmstore.EmptyCode) {
			

		}

		// contract account
		if daRootHash != types.EmptyRootHash && !bytes.Equal(da.CodeHash, evmstore.EmptyCode){
			stIt, err := snaptree.StorageIterator(root, accIt.Hash(), common.Hash{})
			if err != nil {
				return err
			}

			for stIt.Next() {
				da.Storage[stIt.Hash()] = common.Bytes2Hex(stIt.Slot())
			}
		}
		accounts++





	}



	return nil
}

// TODO rewrite it using c.RWCursor(PlainState) its faster
func writeAccountData(db kv.RwDB, acc eaccounts.Account, addr ecommon.Address) error {
	return db.Update(context.Background(), func(tx kv.RwTx) error {
		value := make([]byte, acc.EncodingLengthForStorage())
		acc.EncodeForStorage(value)
		return tx.Put(kv.PlainState, addr[:], value)
	})
}

// TODO add incarnation to contract based accounts
// TODO fix contract sotrage
// ask about how to write in more efficient way using RwCursor 
func writeAccountStorage(db kv.RwDB, acc eaccounts.Account, addr ecommon.Address) error {
	return db.Update(context.Background(), func(tx kv.RwTx) error {
		compositeKey := dbutils.PlainGenerateCompositeStorageKey(addr.Bytes(), acc.Incarnation, acc.Root.Bytes())
		value := acc.CodeHash.Bytes()
		return tx.Put(kv.PlainState, compositeKey, value)
	})
}

func transformStateAccount(account state.Account) (eAccount eaccounts.Account) {
	bal, overflow := uint256.FromBig(account.Balance)
	if overflow {
		panic(fmt.Sprintf("overflow occured while converting from account.Balance"))
	}
	eAccount.Nonce = account.Nonce
	eAccount.Balance = *bal
	eAccount.Root = ecommon.Hash(account.Root)
	eAccount.CodeHash = crypto.Keccak256Hash(account.CodeHash)

	return 
}


