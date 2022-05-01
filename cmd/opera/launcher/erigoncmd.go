package launcher

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path"
	"path/filepath"
	"time"

	//"fmt"

	"gopkg.in/urfave/cli.v1"
	//"github.com/holiman/uint256"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	//"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"

	//"github.com/ethereum/go-ethereum/crypto"

	"github.com/Fantom-foundation/go-opera/gossip/evmstore"
	"github.com/Fantom-foundation/go-opera/integration"

	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/erigon-lib/kv/mdbx"
	elog "github.com/ledgerwatch/log/v3"
	//estate "github.com/ledgerwatch/erigon/core/state"
	//eaccounts "github.com/ledgerwatch/erigon/core/types/accounts"
	//ecommon "github.com/ledgerwatch/erigon/common"
	//"github.com/ledgerwatch/erigon/crypto"
)

/*
Data model
TBD use state.Account instead of erigon accounts.Account (have different format)

kv.Plainstate
erigon
key - address (unhashed)
value - account encoded for storage
my impl
[]byte          
key account.Hash 
value account in bytes (TODO encoding for storage)

erigon 
kv.HashedAccounts
// key - address hash
// value - account encoded for storage
kv.HashedStorage
//key - address hash + incarnation + storage key hash
//value - storage value(common.hash)

TrieOfAccounts and TrieOfStorage
hasState,groups - mark prefixes existing in hashed_account table
hasTree - mark prefixes existing in trie_account table (not related with branchNodes)
hasHash - mark prefixes which hashes are saved in current trie_account record (actually only hashes of branchNodes can be saved)
@see UnmarshalTrieNode
@see integrity.Trie

+-----------------------------------------------------------------------------------------------------+
| DB record: 0x0B, hasState: 0b1011, hasTree: 0b1001, hasHash: 0b1001, hashes: [x,x]                  |
+-----------------------------------------------------------------------------------------------------+
                |                                           |                               |
                v                                           |                               v
+---------------------------------------------+             |            +--------------------------------------+
| DB record: 0x0B00, hasState: 0b10001        |             |            | DB record: 0x0B03, hasState: 0b10010 |
| hasTree: 0, hasHash: 0b10000, hashes: [x]   |             |            | hasTree: 0, hasHash: 0, hashes: []   |
+---------------------------------------------+             |            +--------------------------------------+
        |                    |                              |                         |                  |
        v                    v                              v                         v                  v
+------------------+    +----------------------+     +---------------+        +---------------+  +---------------+
| Account:         |    | BranchNode: 0x0B0004 |     | Account:      |        | Account:      |  | Account:      |
| 0x0B0000...      |    | has no record in     |     | 0x0B01...     |        | 0x0B0301...   |  | 0x0B0304...   |
| in HashedAccount |    |     TrieAccount      |     |               |        |               |  |               |
+------------------+    +----------------------+     +---------------+        +---------------+  +---------------+
                           |                |
                           v                v
		           +---------------+  +---------------+
		           | Account:      |  | Account:      |
		           | 0x0B000400... |  | 0x0B000401... |
		           +---------------+  +---------------+
Invariants:
- hasTree is subset of hasState
- hasHash is subset of hasState
- first level in account_trie always exists if hasState>0
- TrieStorage record of account.root (length=40) must have +1 hash - it's account.root
- each record in TrieAccount table must have parent (may be not direct) and this parent must have correct bit in hasTree bitmap
- if hasState has bit - then HashedAccount table must have record according to this bit
- each TrieAccount record must cover some state (means hasState is always > 0)
- TrieAccount records with length=1 can satisfy (hasBranch==0&&hasHash==0) condition
- Other records in TrieAccount and TrieStorage must (hasTree!=0 || hasHash!=0)



erigon
Physical layout:
PlainState and HashedStorage utilises DupSort feature of MDBX (store multiple values inside 1 key).
-------------------------------------------------------------
	   key              |            value
-------------------------------------------------------------
[acc_hash]              | [acc_value]
[acc_hash]+[inc]        | [storage1_hash]+[storage1_value]
                        | [storage2_hash]+[storage2_value] // this value has no own key. it's 2nd value of [acc_hash]+[inc] key.
                        | [storage3_hash]+[storage3_value]
                        | ...
[acc_hash]+[old_inc]    | [storage1_hash]+[storage1_value]
                        | ...
[acc2_hash]             | [acc2_value]


*/

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

	accKV, err := mdbx.NewMDBX(elog.New()).
			Path(filepath.Join(os.TempDir(), "lmdb")).
			WithTablessCfg(func(defaultBuckets kv.TableCfg) kv.TableCfg {
				return kv.TableCfg{
					kv.PlainState: kv.TableCfgItem{},
					kv.Code: kv.TableCfgItem{},
		            }
			}).Open()
	if err != nil {
		return err
	}
	defer accKV.Close()


	/*
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
	*/

	/*
	TODO fill in
	kv.HashedAccounts, 
	hash1 := common.HexToHash("0x30af561000000000000000000000000000000000000000000000000000000000")
	assert.Nil(t, tx.Put(kv.HashedAccounts, hash1[:], encoded))



	kv.TrieOfAccounts, 
	kv.TrieOfStorage, 
	kv.HashedStorage
    tables to call CalcTrieRoot to compute state root
	https://github.com/ledgerwatch/erigon/blob/devel/turbo/trie/trie_root.go#L197


	c, err := tx.Cursor(kv.PlainState)
			if err != nil {
				return nil, nil, err
			}
			h := common.NewHasher()
			defer common.ReturnHasherToPool(h)
			for k, v, err := c.First(); k != nil; k, v, err = c.Next() {
				if err != nil {
					return nil, nil, fmt.Errorf("interate over plain state: %w", err)
				}
				var newK []byte
				if len(k) == common.AddressLength {
					newK = make([]byte, common.HashLength)
				} else {
					newK = make([]byte, common.HashLength*2+common.IncarnationLength)
				}
				h.Sha.Reset()
				//nolint:errcheck
				h.Sha.Write(k[:common.AddressLength])
				//nolint:errcheck
				h.Sha.Read(newK[:common.HashLength])
				if len(k) > common.AddressLength {
					copy(newK[common.HashLength:], k[common.AddressLength:common.AddressLength+common.IncarnationLength])
					h.Sha.Reset()
					//nolint:errcheck
					h.Sha.Write(k[common.AddressLength+common.IncarnationLength:])
					//nolint:errcheck
					h.Sha.Read(newK[common.HashLength+common.IncarnationLength:])
					if err = tx.Put(kv.HashedStorage, newK, common.CopyBytes(v)); err != nil {
						return nil, nil, fmt.Errorf("insert hashed key: %w", err)
					}
				} else {
					if err = tx.Put(kv.HashedAccounts, newK, common.CopyBytes(v)); err != nil {
						return nil, nil, fmt.Errorf("insert hashed key: %w", err)
					}
				}

			}


	*/

	for accIter.Next() {
		accounts += 1
		var acc state.Account

		if err := rlp.DecodeBytes(accIter.Value, &acc); err != nil {
			log.Error("Invalid account encountered during traversal", "err", err)
			return err
		}

		// Writing account into LMDB kv.Plainstate table
		if err := accKV.Update(context.Background(), func(tx kv.RwTx) error {
			// TODO add EncodeForStorage for state.Account
			key := common.BytesToHash(accIter.Value)
			return tx.Put(kv.PlainState, key.Bytes(), accIter.Value)
		}); err != nil {
			return err
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
			// Writing contract code into LMDB kv.Code table
			if err := accKV.Update(context.Background(), func(tx kv.RwTx) error {
				//eAccount := transformAccToErigon(acc)
				//value := make([]byte, eAccount.EncodingLengthForStorage())
				//eAccount.EncodeForStorage(value)
				key := common.BytesToHash(acc.CodeHash)
				return tx.Put(kv.Code, key.Bytes(), acc.CodeHash[:])
			}); err != nil {
				return err
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

