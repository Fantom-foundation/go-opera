package erigon

import (
	"fmt"
	"math/big"
	"math/rand"
	"regexp"
	"testing"

	"github.com/holiman/uint256"

	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/state/snapshot"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethdb/memorydb"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"

	ecommon "github.com/ledgerwatch/erigon/common"
	"github.com/ledgerwatch/erigon/core/types/accounts"
	eaccounts "github.com/ledgerwatch/erigon/core/types/accounts"
)

var (
	// emptyRoot is the known root hash of an empty trie.
	emptyRoot = common.HexToHash("56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421")

	// emptyCode is the known hash of the empty EVM bytecode.
	//emptyCode = crypto.Keccak256Hash(nil)

)

/*  TestPlan
   1. generate few state accounts ( contract and non contract, also with with empty root and non empty codehash)
   2. put the storage into account
   2. add them into snapshort trie and generate root hash of a trie
   3. accounts from step 1 transform to erigon format
   4. write these accounts into PlainState, Hashstate, intermediate state to compute root hash
   5. comapre two root hashes from step 2 and 4




func randomAccount(t *testing.T) (*accounts.Account, common.Address) {
	t.Helper()
	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	acc := accounts.NewAccount()
	acc.Initialised = true
	acc.Balance = *uint256.NewInt(uint64(rand.Int63()))
	addr := crypto.PubkeyToAddress(key.PublicKey)
	return &acc, addr
}

func generateAccountsWithStorageAndHistory(t *testing.T, blockWriter *PlainStateWriter, numOfAccounts, numOfStateKeys int) ([]common.Address, []*accounts.Account, []map[common.Hash]uint256.Int, []*accounts.Account, []map[common.Hash]uint256.Int) {
	t.Helper()

	accHistory := make([]*accounts.Account, numOfAccounts)
	accState := make([]*accounts.Account, numOfAccounts)
	accStateStorage := make([]map[common.Hash]uint256.Int, numOfAccounts)
	accHistoryStateStorage := make([]map[common.Hash]uint256.Int, numOfAccounts)
	addrs := make([]common.Address, numOfAccounts)
	for i := range accHistory {
		accHistory[i], addrs[i] = randomAccount(t)
		accHistory[i].Balance = *uint256.NewInt(100)
		accHistory[i].CodeHash = common.Hash{uint8(10 + i)}
		accHistory[i].Root = common.Hash{uint8(10 + i)}
		accHistory[i].Incarnation = uint64(i + 1)

		accState[i] = accHistory[i].SelfCopy()
		accState[i].Nonce++
		accState[i].Balance = *uint256.NewInt(200)

		accStateStorage[i] = make(map[common.Hash]uint256.Int)
		accHistoryStateStorage[i] = make(map[common.Hash]uint256.Int)
		for j := 0; j < numOfStateKeys; j++ {
			key := common.Hash{uint8(i*100 + j)}
			newValue := uint256.NewInt(uint64(j))
			if !newValue.IsZero() {
				// Empty value is not considered to be present
				accStateStorage[i][key] = *newValue
			}

			value := uint256.NewInt(uint64(10 + j))
			accHistoryStateStorage[i][key] = *value
			if err := blockWriter.WriteAccountStorage(addrs[i], accHistory[i].Incarnation, &key, value, newValue); err != nil {
				t.Fatal(err)
			}
		}
		if err := blockWriter.UpdateAccountData(addrs[i], accHistory[i], accState[i]); err != nil {
			t.Fatal(err)
		}
	}
	if err := blockWriter.WriteChangeSets(); err != nil {
		t.Fatal(err)
	}
	if err := blockWriter.WriteHistory(); err != nil {
		t.Fatal(err)
	}
	return addrs, accState, accStateStorage, accHistory, accHistoryStateStorage
}





   type storageData struct {
	addr   common.Address
	inc    uint64
	key    common.Hash
	oldVal *uint256.Int
	newVal *uint256.Int
}

func writeStorageBlockData(t *testing.T, blockWriter *PlainStateWriter, data []storageData) {

	for i := range data {
		if err := blockWriter.WriteAccountStorage(data[i].addr, data[i].inc, &data[i].key, data[i].oldVal, data[i].newVal); err != nil {
			t.Fatal(err)
		}
	}

	if err := blockWriter.WriteChangeSets(); err != nil {
		t.Fatal(err)
	}
	if err := blockWriter.WriteHistory(); err != nil {
		t.Fatal(err)
	}
}

// go-ethereum storage.Slot is map[common.Hash][]byte
func copyStorage(storage map[common.Hash]map[common.Hash][]byte) map[common.Hash]map[common.Hash][]byte {
	copy := make(map[common.Hash]map[common.Hash][]byte)
	for accHash, slots := range storage {
		copy[accHash] = make(map[common.Hash][]byte)
		for slotHash, blob := range slots {
			copy[accHash][slotHash] = blob
		}
	}
	return copy
}


func TestDiskSeek(t *testing.T) {
	// Create some accounts in the disk layer
	var db ethdb.Database

	if dir, err := ioutil.TempDir("", "disklayer-test"); err != nil {
		t.Fatal(err)
	} else {
		defer os.RemoveAll(dir)
		diskdb, err := leveldb.New(dir, 256, 0, "", false)
		if err != nil {
			t.Fatal(err)
		}
		db = rawdb.NewDatabase(diskdb)
	}
	// Fill even keys [0,2,4...]
	for i := 0; i < 0xff; i += 2 {
		acc := common.Hash{byte(i)}
		rawdb.WriteAccountSnapshot(db, acc, acc[:])
	}
	// Add an 'higher' key, with incorrect (higher) prefix
	highKey := []byte{rawdb.SnapshotAccountPrefix[0] + 1}
	db.Put(highKey, []byte{0xff, 0xff})

	baseRoot := randomHash()
	rawdb.WriteSnapshotRoot(db, baseRoot)

	snaps := &Tree{
		layers: map[common.Hash]snapshot{
			baseRoot: &diskLayer{
				diskdb: db,
				cache:  fastcache.New(500 * 1024),
				root:   baseRoot,
			},
		},
	}
	// Test some different seek positions
	type testcase struct {
		pos    byte
		expkey byte
	}
	var cases = []testcase{
		{0xff, 0x55}, // this should exit immediately without checking key
		{0x01, 0x02},
		{0xfe, 0xfe},
		{0xfd, 0xfe},
		{0x00, 0x00},
	}
	for i, tc := range cases {
		it, err := snaps.AccountIterator(baseRoot, common.Hash{tc.pos})
		if err != nil {
			t.Fatalf("case %d, error: %v", i, err)
		}
		count := 0
		for it.Next() {
			k, v, err := it.Hash()[0], it.Account()[0], it.Error()
			if err != nil {
				t.Fatalf("test %d, item %d, error: %v", i, count, err)
			}
			// First item in iterator should have the expected key
			if count == 0 && k != tc.expkey {
				t.Fatalf("test %d, item %d, got %v exp %v", i, count, k, tc.expkey)
			}
			count++
			if v != k {
				t.Fatalf("test %d, item %d, value wrong, got %v exp %v", i, count, v, k)
			}
		}
	}
}

snapRoot, err := generateTrieRoot(nil, accIt, common.Hash{}, stackTrieGenerate,
		func(db ethdb.KeyValueWriter, accountHash, codeHash common.Hash, stat *generateStats) (common.Hash, error) {
			storageIt, _ := snap.StorageIterator(accountHash, common.Hash{})
			defer storageIt.Release()

			hash, err := generateTrieRoot(nil, storageIt, accountHash, stackTrieGenerate, nil, stat, false)
			if err != nil {
				return common.Hash{}, err
			}
			return hash, nil
		}, newGenerateStats(), true)


func TestGenerateExistentState(t *testing.T) {
	// We can't use statedb to make a test trie (circular dependency), so make
	// a fake one manually. We're going with a small account trie of 3 accounts,
	// two of which also has the same 3-slot storage trie attached.
	var (
		diskdb = memorydb.New()
		triedb = trie.NewDatabase(diskdb)
	)
	stTrie, _ := trie.NewSecure(common.Hash{}, triedb)
	stTrie.Update([]byte("key-1"), []byte("val-1")) // 0x1314700b81afc49f94db3623ef1df38f3ed18b73a1b7ea2f6c095118cf6118a0
	stTrie.Update([]byte("key-2"), []byte("val-2")) // 0x18a0f4d79cff4459642dd7604f303886ad9d77c30cf3d7d7cedb3a693ab6d371
	stTrie.Update([]byte("key-3"), []byte("val-3")) // 0x51c71a47af0695957647fb68766d0becee77e953df17c29b3c2f25436f055c78
	stTrie.Commit(nil)                              // Root: 0xddefcd9376dd029653ef384bd2f0a126bb755fe84fdcc9e7cf421ba454f2bc67

	accTrie, _ := trie.NewSecure(common.Hash{}, triedb)
	acc := &Account{Balance: big.NewInt(1), Root: stTrie.Hash().Bytes(), CodeHash: emptyCode.Bytes()}
	val, _ := rlp.EncodeToBytes(acc)
	accTrie.Update([]byte("acc-1"), val) // 0x9250573b9c18c664139f3b6a7a8081b7d8f8916a8fcc5d94feec6c29f5fd4e9e
	rawdb.WriteAccountSnapshot(diskdb, hashData([]byte("acc-1")), val)
	rawdb.WriteStorageSnapshot(diskdb, hashData([]byte("acc-1")), hashData([]byte("key-1")), []byte("val-1"))
	rawdb.WriteStorageSnapshot(diskdb, hashData([]byte("acc-1")), hashData([]byte("key-2")), []byte("val-2"))
	rawdb.WriteStorageSnapshot(diskdb, hashData([]byte("acc-1")), hashData([]byte("key-3")), []byte("val-3"))

	acc = &Account{Balance: big.NewInt(2), Root: emptyRoot.Bytes(), CodeHash: emptyCode.Bytes()}
	val, _ = rlp.EncodeToBytes(acc)
	accTrie.Update([]byte("acc-2"), val) // 0x65145f923027566669a1ae5ccac66f945b55ff6eaeb17d2ea8e048b7d381f2d7
	diskdb.Put(hashData([]byte("acc-2")).Bytes(), val)
	rawdb.WriteAccountSnapshot(diskdb, hashData([]byte("acc-2")), val)

	acc = &Account{Balance: big.NewInt(3), Root: stTrie.Hash().Bytes(), CodeHash: emptyCode.Bytes()}
	val, _ = rlp.EncodeToBytes(acc)
	accTrie.Update([]byte("acc-3"), val) // 0x50815097425d000edfc8b3a4a13e175fc2bdcfee8bdfbf2d1ff61041d3c235b2
	rawdb.WriteAccountSnapshot(diskdb, hashData([]byte("acc-3")), val)
	rawdb.WriteStorageSnapshot(diskdb, hashData([]byte("acc-3")), hashData([]byte("key-1")), []byte("val-1"))
	rawdb.WriteStorageSnapshot(diskdb, hashData([]byte("acc-3")), hashData([]byte("key-2")), []byte("val-2"))
	rawdb.WriteStorageSnapshot(diskdb, hashData([]byte("acc-3")), hashData([]byte("key-3")), []byte("val-3"))

	root, _ := accTrie.Commit(nil) // Root: 0xe3712f1a226f3782caca78ca770ccc19ee000552813a9f59d479f8611db9b1fd
	triedb.Commit(root, false, nil)

	snap := generateSnapshot(diskdb, triedb, 16, root)
	select {
	case <-snap.genPending:
		// Snapshot generation succeeded

	case <-time.After(3 * time.Second):
		t.Errorf("Snapshot generation failed")
	}
	checkSnapRoot(t, snap, root)
	// Signal abortion to the generator and wait for it to tear down
	stop := make(chan *generatorStats)
	snap.genAbort <- stop
	<-stop
}
*/

func randomStateAccount(t *testing.T) (state.Account, common.Address) {
	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}

	var acc state.Account
	acc.Balance = big.NewInt(int64(rand.Int63()))
	acc.Root = common.HexToHash("56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421")
	acc.CodeHash = crypto.Keccak256(nil)
	addr := crypto.PubkeyToAddress(key.PublicKey)

	return acc, addr
}

func generateStateAccountsWithStorage(t *testing.T, numAccounts int) ([]common.Address, []state.Account, []map[common.Hash]string) {
	accounts := make([]state.Account, numAccounts)
	accStorage := make([]map[common.Hash]string, numAccounts)
	addr := make([]common.Address, numAccounts)
	for i := 0; i < numAccounts; i++ {
		accounts[i], addr[i] = randomStateAccount(t)
		key, value := common.Hash{uint8(i * 100)}, addr[i].Hex()
		accStorage[i][key] = value
	}

	return addr, accounts, accStorage
}

/*
func makeSnapTree()  {
	snaptree, err := snapshot.New(diskdb, trie.NewDatabase(diskdb), 256, root, false, false, false)
	if err != nil {
		return err
	}

}
*/

func TestCompareLegacyErigonStateRoots(t *testing.T) {
	var (
		diskdb = memorydb.New()
		triedb = trie.NewDatabase(diskdb)
	)

	stTrie, _ := trie.NewSecure(common.Hash{}, triedb)
	stTrie.Update([]byte("key-1"), []byte("val-1")) // 0x1314700b81afc49f94db3623ef1df38f3ed18b73a1b7ea2f6c095118cf6118a0
	stTrie.Update([]byte("key-2"), []byte("val-2")) // 0x18a0f4d79cff4459642dd7604f303886ad9d77c30cf3d7d7cedb3a693ab6d371
	stTrie.Update([]byte("key-3"), []byte("val-3")) // 0x51c71a47af0695957647fb68766d0becee77e953df17c29b3c2f25436f055c78
	stRoot, err := stTrie.Commit(nil)               // Root: 0xddefcd9376dd029653ef384bd2f0a126bb755fe84fdcc9e7cf421ba454f2bc67
	require.NoError(t, err)

	stMap := make(map[string]string)
	stMap["key-1"] = "val-1"
	stMap["key-2"] = "val-2"
	stMap["key-3"] = "val-3"

	accMap := make(map[string]*snapshot.Account)
	accTrie, _ := trie.NewSecure(common.Hash{}, triedb)
	acc := &snapshot.Account{Balance: big.NewInt(1), Root: stTrie.Hash().Bytes(), CodeHash: emptyCode}
	_ = transformSnapAccount(acc, false)

	val, _ := rlp.EncodeToBytes(acc)
	accTrie.Update([]byte("acc-1"), val) // 0x9250573b9c18c664139f3b6a7a8081b7d8f8916a8fcc5d94feec6c29f5fd4e9e
	accMap[string(val)] = acc
	rawdb.WriteAccountSnapshot(diskdb, hashData([]byte("acc-1")), val)
	rawdb.WriteStorageSnapshot(diskdb, hashData([]byte("acc-1")), hashData([]byte("key-1")), []byte("val-1"))
	rawdb.WriteStorageSnapshot(diskdb, hashData([]byte("acc-1")), hashData([]byte("key-2")), []byte("val-2"))
	rawdb.WriteStorageSnapshot(diskdb, hashData([]byte("acc-1")), hashData([]byte("key-3")), []byte("val-3"))

	acc = &snapshot.Account{Balance: big.NewInt(2), Root: emptyRoot.Bytes(), CodeHash: emptyCode}
	val, _ = rlp.EncodeToBytes(acc)
	t.Log("hashData([]byte(val))", hashData([]byte(val)))
	accTrie.Update([]byte("acc-2"), val) // 0x65145f923027566669a1ae5ccac66f945b55ff6eaeb17d2ea8e048b7d381f2d7
	accMap[string(val)] = acc
	diskdb.Put(hashData([]byte("acc-2")).Bytes(), val)
	rawdb.WriteAccountSnapshot(diskdb, hashData([]byte("acc-2")), val)

	acc = &snapshot.Account{Balance: big.NewInt(3), Root: stTrie.Hash().Bytes(), CodeHash: emptyCode}
	val, _ = rlp.EncodeToBytes(acc)
	accTrie.Update([]byte("acc-3"), val) // 0x50815097425d000edfc8b3a4a13e175fc2bdcfee8bdfbf2d1ff61041d3c235b2
	accMap[string(val)] = acc
	rawdb.WriteAccountSnapshot(diskdb, hashData([]byte("acc-3")), val)
	rawdb.WriteStorageSnapshot(diskdb, hashData([]byte("acc-3")), hashData([]byte("key-1")), []byte("val-1"))
	rawdb.WriteStorageSnapshot(diskdb, hashData([]byte("acc-3")), hashData([]byte("key-2")), []byte("val-2"))
	rawdb.WriteStorageSnapshot(diskdb, hashData([]byte("acc-3")), hashData([]byte("key-3")), []byte("val-3"))

	root, _ := accTrie.Commit(nil) // Root: 0xe3712f1a226f3782caca78ca770ccc19ee000552813a9f59d479f8611db9b1fd
	triedb.Commit(root, false, nil)

	//rawdb.WriteSnapshotRoot(diskdb, root)
	generateSnapshot(diskdb, triedb, 16, root)
	snaptree, err := snapshot.New(diskdb, trie.NewDatabase(diskdb), 256, root, false, false, false)
	require.NoError(t, err)

	accIt, err := snaptree.AccountIterator(root, common.Hash{})
	require.NoError(t, err)
	defer accIt.Release()

	require.Equal(t, stRoot.Hex(), stTrie.Hash().Hex())

	for accIt.Next() {
		addr := ecommon.BytesToAddress(accIt.Hash().Bytes()).Hex()
		t.Log("addr", addr)
		require.True(t, isValidAddress(addr))
		t.Log("accIt.Hash()", accIt.Hash())
		t.Log("accIt.Account().Hash()", hashData(accIt.Account()))
		key := string(accIt.Account())
		_, ok := accMap[key]
		require.True(t, ok)

		/*
			if bytes.Equal(acc.Root, stTrie.Hash().Bytes()) {

				stIt, err := snaptree.StorageIterator(root, accIt.Hash(), common.Hash{})
				//require.Equal(t, accIt.Hash().Hex(), "0x18a0f4d79cff4459642dd7604f303886ad9d77c30cf3d7d7cedb3a693ab6d371")
				require.NoError(t, err)
				defer stIt.Release()
				for stIt.Next() {
					t.Log("stIt.Hash()", stIt.Hash())
					t.Log("hashData([]byte(key-2))", hashData([]byte("key-2")))
					t.Log("hashData([]byte(key-2)).Bytes()", string(hashData([]byte("key-2")).Bytes()))
					t.Log("hashData([]byte(val-2))", hashData([]byte("val-2")))
					t.Log("stIt.Hash().String()", stIt.Hash())
					t.Log("stIt.Slot().String()", string(stIt.Slot()))
					val2, ok := stMap["key-2"]
					require.True(t, ok)
					require.Equal(t, val2, string(stIt.Slot()))
					require.Equal(t, hashData(stIt.Slot()), stIt.Hash())
				}
			}
		*/
	}

	/* TODO
	transform every snapshot.Acc to erigon

	write every account to plain.state, Hash.state, intemediate hases




	*/

}

//base := generateSnapshot(diskdb, triedb, 16, root)
//_ = &snapshot.Tree{
//	layers: map[common.Hash]snapshot{
//		base.root: base,
//	},
//}

// TestStorageIteratorTraversalValues
// TestGenerateExistentStateWithWrongStorage
// TestGenerateExistentStateWithWrongAccounts(t *testing.T) {

func transformSnapAccount(account *snapshot.Account, isContractAcc bool) eaccounts.Account {
	eAccount := accounts.NewAccount()
	eAccount.Initialised = true // ?
	bal, overflow := uint256.FromBig(account.Balance)
	if overflow {
		panic(fmt.Sprintf("overflow occured while converting from account.Balance"))
	}
	eAccount.Nonce = account.Nonce
	eAccount.Balance = *bal
	eAccount.Root = ecommon.BytesToHash(account.Root) //?
	eAccount.CodeHash = ecommon.BytesToHash(account.CodeHash)
	if isContractAcc {
		eAccount.Incarnation = 1
	}
	return eAccount
}

func isValidAddress(v string) bool {
	re := regexp.MustCompile("^0x[0-9a-fA-F]{40}$")
	return re.MatchString(v)
}
