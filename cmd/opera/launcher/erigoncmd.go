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

	"gopkg.in/urfave/cli.v1"
	"github.com/holiman/uint256"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	//"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"


	"github.com/Fantom-foundation/go-opera/gossip/evmstore"
	"github.com/Fantom-foundation/go-opera/integration"

	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/erigon-lib/kv/mdbx"

	elog "github.com/ledgerwatch/log/v3"
	//estate "github.com/ledgerwatch/erigon/core/state"
	eaccounts "github.com/ledgerwatch/erigon/core/types/accounts"
	ecommon "github.com/ledgerwatch/erigon/common"
	"github.com/ledgerwatch/erigon/common/dbutils"
	"github.com/ledgerwatch/erigon/crypto"
)



func erigon(ctx *cli.Context) error {

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

	db, err := mdbx.NewMDBX(elog.New()).
			Path(filepath.Join(os.TempDir(), "lmdb")).
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


  
	transformStateAccount := func(account state.Account) (eAccount eaccounts.Account) {
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

	// TODO rewrite it using c.RWCursor(PlainState) its faster
    writeAccountData := func(db kv.RwDB, acc eaccounts.Account, addr ecommon.Address) error {
		return db.Update(context.Background(), func(tx kv.RwTx) error {
			value := make([]byte, acc.EncodingLengthForStorage())
			acc.EncodeForStorage(value)
			return tx.Put(kv.PlainState, addr[:], value)
		})
	}

	// ask about how to write more efficient using Cursor or tx.Put
	writeAccountStorage := func(db kv.RwDB, acc eaccounts.Account, addr ecommon.Address) error {
		return db.Update(context.Background(), func(tx kv.RwTx) error {
			compositeKey := dbutils.PlainGenerateCompositeStorageKey(addr.Bytes(), acc.Incarnation, acc.Root.Bytes())
			value := acc.CodeHash.Bytes()
			return tx.Put(kv.PlainState, compositeKey, value)
		})
	}

	randomAddr := func() ecommon.Address {
		key, _ := crypto.GenerateKey()
		addr := ecommon.Address(crypto.PubkeyToAddress(key.PublicKey))
		return addr
	}



	/*
	TODO fill in
	kv.HashedAccounts, 
	hash1 := common.HexToHash("0x30af561000000000000000000000000000000000000000000000000000000000")
	assert.Nil(t, tx.Put(kv.HashedAccounts, hash1[:], encoded))



    erigon.RegenerateIntermediateHashes (from first block)

	kv.TrieOfAccounts, 
	kv.TrieOfStorage, 
	kv.HashedStorage
	kv.HashedAccounts

    Plan
	0. PlainState
	func (w *PlainStateWriter) UpdateAccountData(address common.Address  
	func (w *PlainStateWriter) WriteAccountStorage(address common.Address)
	1. load all accounts into buckets kv.HashedAcMadrid to delete themcounts, kv.HashedStorage using HashState stage

    


	HashStorage
	incarnation := uint64(1)
	hash3 := common.HexToHash("0xB041000000000000000000000000000000000000000000000000000000000000")
	assert.Nil(t, addTestAccount(tx, hash3, 2*params.Ether, incarnation))

	loc1 := common.HexToHash("0x1200000000000000000000000000000000000000000000000000000000000000")
	loc2 := common.HexToHash("0x1400000000000000000000000000000000000000000000000000000000000000")
	loc3 := common.HexToHash("0x3000000000000000000000000000000000000000000000000000000000E00000")
	loc4 := common.HexToHash("0x3000000000000000000000000000000000000000000000000000000000E00001")

	val1 := common.FromHex("0x42")
	val2 := common.FromHex("0x01")
	val3 := common.FromHex("0x127a89")
	val4 := common.FromHex("0x05")

	assert.Nil(t, tx.Put(kv.HashedStorage, dbutils.GenerateCompositeStorageKey(hash3, incarnation, loc1), val1))
	2. transform state.Account to erigon account or not +

	3. find out where to take 


	erigon.incrementIntermediateHashes
	additionally to above buckers
	kv.Plainstate
	kv.AccountChangeset, 
	kv.StorageChangeSet
*/



	for accIter.Next() {
		accounts += 1
		var acc state.Account

		if err := rlp.DecodeBytes(accIter.Value, &acc); err != nil {
			log.Error("Invalid account encountered during traversal", "err", err)
			return err
		}

		eAccount := transformStateAccount(acc)
		// TODO replace random accs 
		addr := randomAddr()

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

		// write to kv.HashedAccounts and kv.HashedStorae
		
    

		




		// TODO consider to update it through IntraBlockState



	
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
	return nil
}
