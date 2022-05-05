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



	transformAccToErigon := func(account state.Account) (eAccount eaccounts.Account) {
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
	1. load all accounts into buckets kv.HashedAccounts, kv.HashedStorage using HashState stage

    


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
	2. transform state.Account to erigon account or not 
	    +    |   -
		       it can lead to to secp256k1 problems (duplicated symbol, cause c libraries have the same namespace)
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

		eAccount := transformAccToErigon(acc)
		addr := randomAddr()


	
		switch {
		// EOA
		case acc.Root == types.EmptyRootHash && bytes.Equal(acc.CodeHash, evmstore.EmptyCode):
			if err := writeAccountData(db, eAccount, addr); err != nil {
				return err
			}
        
		// contract account
		case acc.Root != types.EmptyRootHash && !bytes.Equal(acc.CodeHash, evmstore.EmptyCode):
			
			if err := writeAccountStorage(db, eAccount, addr, acc.CodeHash); err != nil {
				return err
			}

		// TODO ADDRESS OTHER CASES 
		}
	
    

		

		// Writing account into LMDB kv.Plainstate table

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


func writeAccountData(db kv.RwDB, acc eaccounts.Account, addr ecommon.Address) error {
	return db.Update(context.Background(), func(tx kv.RwTx) error {
		value := make([]byte, acc.EncodingLengthForStorage())
		acc.EncodeForStorage(value)
		return tx.Put(kv.PlainState, addr[:], value)
	})
}


func writeAccountStorage(db kv.RwDB, acc eaccounts.Account, addr ecommon.Address, value []byte ) error {
	compositeKey := dbutils.PlainGenerateCompositeStorageKey(addr.Bytes(), acc.Incarnation, acc.Root.Bytes())
	return db.Update(context.Background(), func(tx kv.RwTx) error {
		return tx.Put(kv.PlainState, compositeKey, value)
	})
}



func randomAddr() ecommon.Address {
	key, _ := crypto.GenerateKey()
	addr := ecommon.Address(crypto.PubkeyToAddress(key.PublicKey))
	return addr
}


/* TODO reimplement it for state.Account
func EncodeForStorage(a state.Account, buffer []byte) {
	var fieldSet = 0 // start with first bit set to 0
	var pos = 1
	if a.Nonce > 0 {
		fieldSet = 1
		nonceBytes := (bits.Len64(a.Nonce) + 7) / 8
		buffer[pos] = byte(nonceBytes)
		var nonce = a.Nonce
		for i := nonceBytes; i > 0; i-- {
			buffer[pos+i] = byte(nonce)
			nonce >>= 8
		}
		pos += nonceBytes + 1
	}

	// Encoding balance
	if !a.Balance.IsZero() {
		fieldSet |= 2
		balanceBytes := a.Balance.ByteLen()
		buffer[pos] = byte(balanceBytes)
		pos++
		a.Balance.WriteToSlice(buffer[pos : pos+balanceBytes])
		pos += balanceBytes
	}

	if a.Incarnation > 0 {
		fieldSet |= 4
		incarnationBytes := (bits.Len64(a.Incarnation) + 7) / 8
		buffer[pos] = byte(incarnationBytes)
		var incarnation = a.Incarnation
		for i := incarnationBytes; i > 0; i-- {
			buffer[pos+i] = byte(incarnation)
			incarnation >>= 8
		}
		pos += incarnationBytes + 1
	}

	// Encoding CodeHash
	if !a.IsEmptyCodeHash() {
		fieldSet |= 8
		buffer[pos] = 32
		copy(buffer[pos+1:], a.CodeHash.Bytes())
		//pos += 33
	}

	buffer[0] = byte(fieldSet)
}
*/

