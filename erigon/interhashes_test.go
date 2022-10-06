package erigon

import (
	"fmt"
	"math/big"
	"math/rand"
	"testing"

	"github.com/Fantom-foundation/go-opera/gossip/evmstore/state/snapshot"
	com "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethdb/memorydb"
	"github.com/ethereum/go-ethereum/rlp"
	legacytrie "github.com/ethereum/go-ethereum/trie"

	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/erigon-lib/kv/memdb"

	"github.com/ledgerwatch/erigon/common"
	account "github.com/ledgerwatch/erigon/core/types/accounts"
	"github.com/ledgerwatch/erigon/crypto"
	etrie "github.com/ledgerwatch/erigon/turbo/trie"

	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
)

var (
	key1 = common.HexToHash("0xB1A0000000000000000000000000000000000000000000000000000000000000")
	key2 = []byte("acc-2")
	key3 = []byte("acc-3")
)

func addSnapTestAccount(balance int64) []byte {
	acc := &snapshot.Account{Balance: big.NewInt(1), Root: emptyRoot.Bytes(), CodeHash: emptyCode}
	val, _ := rlp.EncodeToBytes(acc)
	return val
}

// address Expected nil, but got: &fmt.wrapError{msg:"fail DecodeForStorage: codehash should be 32 bytes long, got 68 instead", err:(*errors.errorString)(0x140003f59e0)}
/*
func TestCompareEthereumErigonStateRootWithSnaphotAccounts(t *testing.T) {
	var (
		diskdb = memorydb.New()
		triedb = trie.NewDatabase(diskdb)
	)
    // compute state root of 3 test snapshot accounts
	accMap := make(map[string][]byte)
	tr, _ := trie.NewSecure(com.Hash{}, triedb)

	val := addSnapTestAccount(1)
	tr.Update(key1, val) // 0xc7a30f39aff471c95d8a837497ad0e49b65be475cc0953540f80cfcdbdcd9074
	accMap[string(key1)] = val

	val = addSnapTestAccount(2)
	tr.Update(key2, val) // 0x65145f923027566669a1ae5ccac66f945b55ff6eaeb17d2ea8e048b7d381f2d7
	accMap[string(key2)] = val

	val = addSnapTestAccount(3)
	tr.Update(key3, val) // 0x19ead688e907b0fab07176120dceec244a72aff2f0aa51e8b827584e378772f4
	accMap[string(key3)] = val

	legacyRoot, err := tr.Commit(nil)         // Root: 0xa04693ea110a31037fb5ee814308a6f1d76bdab0b11676bdf4541d2de55ba978
    assert.NoError(t, err)
	assert.Equal(t, legacyRoot.Hex(), "0xa04693ea110a31037fb5ee814308a6f1d76bdab0b11676bdf4541d2de55ba978")


	// setup erigon test tx
	_, tx := memdb.NewTestTx(t)

	// insert byte representation of 3 snapshot accounts into kv.HashedAccounts bucket
	assert.Nil(t, tx.Put(kv.HashedAccounts, key1, accMap[string(key1)]))
	assert.Nil(t, tx.Put(kv.HashedAccounts, key2, accMap[string(key2)]))
	assert.Nil(t, tx.Put(kv.HashedAccounts, key3, accMap[string(key3)]))




	/*
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
	assert.Nil(t, tx.Put(kv.HashedStorage, dbutils.GenerateCompositeStorageKey(hash3, incarnation, loc2), val2))
	assert.Nil(t, tx.Put(kv.HashedStorage, dbutils.GenerateCompositeStorageKey(hash3, incarnation, loc3), val3))
	assert.Nil(t, tx.Put(kv.HashedStorage, dbutils.GenerateCompositeStorageKey(hash3, incarnation, loc4), val4))

	hash4a := common.HexToHash("0xB1A0000000000000000000000000000000000000000000000000000000000000")
	assert.Nil(t, addTestAccount(tx, hash4a, 4*params.Ether, 0))

	hash5 := common.HexToHash("0xB310000000000000000000000000000000000000000000000000000000000000")
	assert.Nil(t, addTestAccount(tx, hash5, 8*params.Ether, 0))

	hash6 := common.HexToHash("0xB340000000000000000000000000000000000000000000000000000000000000")
	assert.Nil(t, addTestAccount(tx, hash6, 1*params.Ether, 0))
*/

// ----------------------------------------------------------------
// Populate account & storage trie DB tables
// ----------------------------------------------------------------

/*
	cfg := StageTrieCfg(nil, false, true, t.TempDir())
	//expHash := common.BytesToHash([]byte("ssjj"))
	hash, err := RegenerateIntermediateHashes("IH", tx, cfg, common.Hash{} /* expectedRootHash */ /*nil*/ /* quit )*/
/*
	assert.NotNil(t, hash.Hex())
	assert.Nil(t, err)
	assert.Equal(t, legacyRoot, hash)
*/

// ----------------------------------------------------------------
// Check account trie
// ----------------------------------------------------------------

/*
	accountTrieA := make(map[string][]byte)
	err = tx.ForEach(kv.TrieOfAccounts, nil, func(k, v []byte) error {
		accountTrieA[string(k)] = common.CopyBytes(v)
		return nil
	})
	assert.Nil(t, err)

	assert.Equal(t, 2, len(accountTrieA)) // error

	/*
	hasState1a, hasTree1a, hasHash1a, hashes1a, rootHash1a := etrie.UnmarshalTrieNode(accountTrieA[string(common.FromHex("0B"))])
	assert.Equal(t, uint16(0b1011), hasState1a)
	assert.Equal(t, uint16(0b0001), hasTree1a)
	assert.Equal(t, uint16(0b1001), hasHash1a)
	assert.Equal(t, 2*length.Hash, len(hashes1a))
	assert.Equal(t, 0, len(rootHash1a))

	hasState2a, hasTree2a, hasHash2a, hashes2a, rootHash2a := etrie.UnmarshalTrieNode(accountTrieA[string(common.FromHex("0B00"))])
	assert.Equal(t, uint16(0b10001), hasState2a)
	assert.Equal(t, uint16(0b00000), hasTree2a)
	assert.Equal(t, uint16(0b10000), hasHash2a)
	assert.Equal(t, 1*length.Hash, len(hashes2a))
	assert.Equal(t, 0, len(rootHash2a))
*/

// ----------------------------------------------------------------
// Check storage trie
// ----------------------------------------------------------------

/*
	storageTrie := make(map[string][]byte)
	err = tx.ForEach(kv.TrieOfStorage, nil, func(k, v []byte) error {
		storageTrie[string(k)] = common.CopyBytes(v)
		return nil
	})
	assert.Nil(t, err)

	assert.Equal(t, 1, len(storageTrie))

	storageKey := make([]byte, length.Hash+8)
	copy(storageKey, hash3.Bytes())
	binary.BigEndian.PutUint64(storageKey[length.Hash:], incarnation)

	hasState3, hasTree3, hasHash3, hashes3, rootHash3 := etrie.UnmarshalTrieNode(storageTrie[string(storageKey)])
	assert.Equal(t, uint16(0b1010), hasState3)
	assert.Equal(t, uint16(0b0000), hasTree3)
	assert.Equal(t, uint16(0b0010), hasHash3)
	assert.Equal(t, 1*length.Hash, len(hashes3))
	assert.Equal(t, length.Hash, len(rootHash3))


}
*/

func addErigonTestAccount(tx kv.Putter, balance uint64) ([]byte, error) {
	acc := account.NewAccount()
	acc.Balance.SetUint64(balance)

	encoded := make([]byte, acc.EncodingLengthForStorage())
	acc.EncodeForStorage(encoded)

	return encoded, tx.Put(kv.HashedAccounts, key1[:], encoded)
}

// erigon calcTrieRoot algorithm serializes a leafnode here, but geth trie algorithm serializes a shortnode.
// Due to this descrepancy, different state roots are computed. To resolve this issue, CalcTrieRoot stare root should be compared with erigon trie state root , not with geth trie root.
// Please see TestErigonTrie3AccsRegenerateIntermediateHashes below for more information
func TestStateRootsNotMatchWithErigonAccounts(t *testing.T) {
	var (
		diskdb = memorydb.New()
		triedb = legacytrie.NewDatabase(diskdb)

		_, tx = memdb.NewTestTx(t)
	)

	// 1.make a tree
	//tr, _ := trie.NewSecure(com.Hash{}, triedb)
	tr, err := legacytrie.New(com.Hash{}, triedb)
	assert.NoError(t, err)

	/*
		hashedKey, err := common.HashData(key1)
		assert.NoError(t, err)
		t.Log("test hashedKey", hashedKey.Hex())
	*/

	/*
		hashedKey, err := common.HashData(key1)
		assert.NoError(t, err)
		t.Log("test hashedKey.Hex(): ", hashedKey.Hex())
	*/

	val, err := addErigonTestAccount(tx, 1)
	assert.Nil(t, err)

	fmt.Printf("hex value of accountKey: %x\n", key1.Bytes())
	fmt.Printf("hex value of serialized account: %x\n", val)

	assert.Nil(t, tr.TryUpdate(key1[:], val))
	//tr.SetRoot()

	/*
		val, err = addErigonTestAccount(tx, 2, key2)
		assert.Nil(t, err)
		tr.Update(key2, val)

		val, err = addErigonTestAccount(tx, 3, key3)
		assert.Nil(t, err)
		tr.Update(key3, val)
	*/

	// generating ethereum state root
	// 0xe1a85473f43bee6e19dc51a178326327eb61edea2fe1ab6cc5b90c814b1eb371
	//legacyRoot, err := tr.Commit(nil)
	assert.NoError(t, err)
	//assert.Equal(t, "0xa04693ea110a31037fb5ee814308a6f1d76bdab0b11676bdf4541d2de55ba978", legacyRoot.Hex())

	//cfg := StageTrieCfg(nil, false, true, t.TempDir())

	//erigonRoot, err := RegenerateIntermediateHashes("IH", tx, cfg, common.Hash{} /* expectedRootHash */, nil /* quit */)

	//legacy: "0xe1a85473f43bee6e19dc51a178326327eb61edea2fe1ab6cc5b90c814b1eb371"
	//erigon  : "0x7ed8e10e694f87e13ac1db95f0ebdea4a4644203edcd6b2b9f6c27e31bf1353f"
	//assert.Equal(t, legacyRoot.Hex(), erigonRoot.Hex())
}

func addErigonTestAccountForStorage(tx kv.Putter, addr common.Address, acc *account.Account) error {
	encoded := make([]byte, acc.EncodingLengthForStorage())
	acc.EncodeForStorage(encoded)
	hash := addr.Hash()
	return tx.Put(kv.HashedAccounts, hash[:], encoded)
}
func TestErigonTrie3AccsRegenerateIntermediateHashes(t *testing.T) {
	_, tx := memdb.NewTestTx(t)

	addr1 := common.HexToAddress("0xa94f5374fce5edbc8e2a8697c15331677e6ebf0b")
	acc1 := &account.Account{
		Nonce:    1,
		Balance:  *uint256.NewInt(209488),
		Root:     common.HexToHash("0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421"),
		CodeHash: common.HexToHash("0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470"),
	}
	assert.Nil(t, addErigonTestAccountForStorage(tx, addr1, acc1))

	addr2 := common.HexToAddress("0xb94f5374fce5edbc8e2a8697c15331677e6ebf0b")
	acc2 := &account.Account{
		Nonce:    0,
		Balance:  *uint256.NewInt(0),
		Root:     common.HexToHash("0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421"),
		CodeHash: common.HexToHash("0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470"),
	}

	assert.Nil(t, addErigonTestAccountForStorage(tx, addr2, acc2))

	addr3 := common.HexToAddress("0xc94f5374fce5edbc8e2a8697c15331677e6ebf0b")
	acc3 := &account.Account{
		Nonce:    0,
		Balance:  *uint256.NewInt(1010),
		Root:     common.HexToHash("0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421"),
		CodeHash: common.HexToHash("0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470"),
	}

	assert.Nil(t, addErigonTestAccountForStorage(tx, addr3, acc3))

	trie := etrie.New(common.Hash{})
	trie.UpdateAccount(addr1.Hash().Bytes(), acc1)
	trie.UpdateAccount(addr2.Hash().Bytes(), acc2)
	trie.UpdateAccount(addr3.Hash().Bytes(), acc3)

	//	trieHash := trie.Hash()

	//	cfg := StageTrieCfg(nil, false, true, t.TempDir())
	//	erigonRoot, err := RegenerateIntermediateHashes("IH", tx, cfg, common.Hash{} /* expectedRootHash */, nil /* quit */)
	//	assert.Nil(t ,err)
	//	assert.Equal(t, trieHash.Hex(), erigonRoot.Hex())
}

func makeAccounts(size int) ([][]byte, []common.Hash) {

	accounts := make([][]byte, size)
	addresses := make([]common.Hash, size)
	for i := 0; i < size; i++ {
		acc := account.NewAccount()
		acc.Initialised = true
		acc.Balance = *uint256.NewInt(uint64(rand.Int63()))

		val := make([]byte, acc.EncodingLengthForStorage())
		acc.EncodeForStorage(val)
		accounts[i] = val

		key, err := crypto.GenerateKey()
		if err != nil {
			panic(err)
		}
		addr := crypto.PubkeyToAddress(key.PublicKey)
		addresses[i] = addr.Hash()
	}

	return accounts, addresses
}

func benchmarkCommitAfterHashFixedSize(b *testing.B, accounts [][]byte, addresses []common.Hash) {
	b.ReportAllocs()
	trie, _ := legacytrie.New(com.Hash{}, legacytrie.NewDatabase(memorydb.New()))
	for i := 0; i < len(addresses); i++ {
		trie.Update(addresses[i][:], accounts[i])
	}
	// Insert the accounts into the trie and hash it
	trie.Hash()
	b.StartTimer()
	trie.Commit(nil)
	b.StopTimer()
}

func BenchmarkCommitAfterHashFixedSize(b *testing.B) {
	b.Run("10", func(b *testing.B) {
		b.StopTimer()
		acc, add := makeAccounts(10)
		for i := 0; i < b.N; i++ {
			benchmarkCommitAfterHashFixedSize(b, acc, add)
		}
	})

	/*
		b.Run("100", func(b *testing.B) {
			b.StopTimer()
			acc, add := makeAccounts(100)
			for i := 0; i < b.N; i++ {
				benchmarkCommitAfterHashFixedSize(b, acc, add)
			}
		})


		b.Run("1K", func(b *testing.B) {
			b.StopTimer()
			acc, add := makeAccounts(1000)
			for i := 0; i < b.N; i++ {
				benchmarkCommitAfterHashFixedSize(b, acc, add)
			}
		})
		b.Run("10K", func(b *testing.B) {
			b.StopTimer()
			acc, add := makeAccounts(10000)
			for i := 0; i < b.N; i++ {
				benchmarkCommitAfterHashFixedSize(b, acc, add)
			}
		})
		b.Run("100K", func(b *testing.B) {
			b.StopTimer()
			acc, add := makeAccounts(100000)
			for i := 0; i < b.N; i++ {
				benchmarkCommitAfterHashFixedSize(b, acc, add)
			}
		})
	*/
}

func benchmarkErigonRegenerateIntermediateHashes(b *testing.B, accounts [][]byte, addresses []common.Hash, tx kv.RwTx) {
	//b.ReportAllocs()
	for i := 0; i < len(addresses); i++ {
		if err := tx.Put(kv.HashedAccounts, addresses[i].Bytes(), accounts[i]); err != nil {
			b.FailNow()
		}

	}

	b.StartTimer()

	cfg := StageTrieCfg(nil, false, true, b.TempDir())
	if _, err := RegenerateIntermediateHashesBench("IH", tx, cfg, common.Hash{} /* expectedRootHash */, nil /* quit */); err != nil {
		b.FailNow()
	}
	b.StopTimer()

}

func BenchmarkErigonRegenerateIntermediateHashes(b *testing.B) {
	b.Run("10", func(b *testing.B) {
		b.StopTimer()
		acc, add := makeAccounts(10)
		_, tx := memdb.NewTestTx(b)
		defer tx.Rollback()
		for i := 0; i < b.N; i++ {
			benchmarkErigonRegenerateIntermediateHashes(b, acc, add, tx)
		}
	})
}

/*
func BenchmarkRegenerateIntermediateHashes(b *testing.B) {

}

func BenchmarkIncrementIntermediateHashes(b *testing.B) {

}
*/

/*
func TestHashTree(t *testing.T) {

	eTrie := etrie.New(common.Hash{})

	acc := new(accounts.Account)
	acc.Root = common.Hash(emptyRoot)
	acc.CodeHash = common.Hash(emptyCode)
	acc.Balance.SetUint64(1)

	eTrie.UpdateAccount(key1[:], acc)
	expHash := eTrie.Hash()


	diskdb := memorydb.New()
	triedb := trie.NewDatabase(diskdb)
	tr, err := trie.New(com.Hash{}, triedb)
	assert.NoError(t, err)

	val := make([]byte, acc.EncodingLengthForStorage())
	acc.EncodeForStorage(val)
	assert.Nil(t, tr.TryUpdate(key1[:], val))
	actualHash, err := tr.Commit(nil)
	assert.NoError(t, err)

	assert.Equal(t, actualHash.Hex(), expHash.Hex())
}
*/

/*
func TestHashWithModificationsNoChanges(t *testing.T) {
	tr := New(common.Hash{})
	// Populate the trie
	var preimage [4]byte
	var keys []string
	for b := uint32(0); b < 10; b++ {
		binary.BigEndian.PutUint32(preimage[:], b)
		key := crypto.Keccak256(preimage[:])
		keys = append(keys, string(key))
	}
	sort.Strings(keys)
	for i, key := range keys {
		if i > 0 && keys[i-1] == key {
			fmt.Printf("Duplicate!\n")
		}
	}
	var a0, a1 accounts.Account
	a0.Balance.SetUint64(100000)
	a0.Root = EmptyRoot
	a0.CodeHash = emptyState
	a0.Initialised = true
	a1.Balance.SetUint64(200000)
	a1.Root = EmptyRoot
	a1.CodeHash = emptyState
	a1.Initialised = true
	v := []byte("VALUE")
	for i, key := range keys {
		if i%2 == 0 {
			tr.UpdateAccount([]byte(key), &a0)
		} else {
			tr.UpdateAccount([]byte(key), &a1)
			// Add storage items too
			for _, storageKey := range keys {
				tr.Update([]byte(key+storageKey), v)
			}
		}
	}
	expectedHash := tr.Hash()
	// Build the root
	var stream Stream
	hb := NewHashBuilder(false)
	rootHash, err := HashWithModifications(
		tr,
		common.Hashes{}, []*accounts.Account{}, [][]byte{},
		common.StorageKeys{}, [][]byte{},
		40,
		&stream, // Streams that will be reused for old and new stream
		hb,      // HashBuilder will be reused
		false,
	)
	if err != nil {
		t.Errorf("Could not compute hash with modification: %v", err)
	}
	if rootHash != expectedHash {
		t.Errorf("Expected %x, got: %x", expectedHash, rootHash)
	}
}

func TestHashWithModificationsChanges(t *testing.T) {
	tr := New(common.Hash{})
	// Populate the trie
	var preimage [4]byte
	var keys []string
	for b := uint32(0); b < 10; b++ {
		binary.BigEndian.PutUint32(preimage[:], b)
		key := crypto.Keccak256(preimage[:])
		keys = append(keys, string(key))
	}
	sort.Strings(keys)
	for i, key := range keys {
		if i > 0 && keys[i-1] == key {
			fmt.Printf("Duplicate!\n")
		}
	}
	var a0, a1 accounts.Account
	a0.Balance.SetUint64(100000)
	a0.Root = EmptyRoot
	a0.CodeHash = emptyState
	a0.Initialised = true
	a1.Balance.SetUint64(200000)
	a1.Root = EmptyRoot
	a1.CodeHash = emptyState
	a1.Initialised = true
	v := []byte("VALUE")
	for i, key := range keys {
		if i%2 == 0 {
			tr.UpdateAccount([]byte(key), &a0)
		} else {
			tr.UpdateAccount([]byte(key), &a1)
			// Add storage items too
			for _, storageKey := range keys {
				tr.Update([]byte(key+storageKey), v)
			}
		}
	}
	tr.Hash()
	// Generate account change
	binary.BigEndian.PutUint32(preimage[:], 5000000)
	var insertKey common.Hash
	copy(insertKey[:], crypto.Keccak256(preimage[:]))
	var insertA accounts.Account
	insertA.Balance.SetUint64(300000)
	insertA.Root = etrie.EmptyRoot
	insertA.CodeHash = emptyState
	insertA.Initialised = true

	// Build the root
	var stream etrie.Stream
	hb := etrie.NewHashBuilder(false)
	rootHash, err := etrie.HashWithModifications(
		tr,
		common.Hashes{insertKey}, []*accounts.Account{&insertA}, [][]byte{nil},
		common.StorageKeys{}, [][]byte{},
		40,
		&stream, // Streams that will be reused for old and new stream
		hb,      // HashBuilder will be reused
		false,
	)
	if err != nil {
		t.Errorf("Could not compute hash with modification: %v", err)
	}
	tr.UpdateAccount(insertKey[:], &insertA)
	expectedHash := tr.Hash()
	if rootHash != expectedHash {
		t.Errorf("Expected %x, got: %x", expectedHash, rootHash)
	}
}

func TestIHCursor(t *testing.T) {
	_, tx := memdb.NewTestTx(t)
	require := require.New(t)
	hash := common.HexToHash(fmt.Sprintf("%064d", 0))

	newV := make([]byte, 0, 1024)
	put := func(ks string, hasState, hasTree, hasHash uint16, hashes []common.Hash) {
		k := common.FromHex(ks)
		integrity.AssertSubset(k, hasTree, hasState)
		integrity.AssertSubset(k, hasHash, hasState)
		_ = tx.Put(kv.TrieOfAccounts, k, common.CopyBytes(trie.MarshalTrieNodeTyped(hasState, hasTree, hasHash, hashes, newV)))
	}

	put("00", 0b0000000000000010, 0b0000000000000000, 0b0000000000000010, []common.Hash{hash})
	put("01", 0b0000000000000111, 0b0000000000000010, 0b0000000000000111, []common.Hash{hash, hash, hash})
	put("0101", 0b0000000000000111, 0b0000000000000000, 0b0000000000000111, []common.Hash{hash, hash, hash})
	put("02", 0b1000000000000000, 0b0000000000000000, 0b1000000000000000, []common.Hash{hash})
	put("03", 0b0000000000000001, 0b0000000000000001, 0b0000000000000000, []common.Hash{})
	put("030000", 0b0000000000000001, 0b0000000000000000, 0b0000000000000001, []common.Hash{hash})
	put("03000e", 0b0000000000000001, 0b0000000000000001, 0b0000000000000001, []common.Hash{hash})
	put("03000e000000", 0b0000000000000100, 0b0000000000000000, 0b0000000000000100, []common.Hash{hash})
	put("03000e00000e", 0b0000000000000100, 0b0000000000000000, 0b0000000000000100, []common.Hash{hash})
	put("05", 0b0000000000000001, 0b0000000000000001, 0b0000000000000001, []common.Hash{hash})
	put("050001", 0b0000000000000001, 0b0000000000000000, 0b0000000000000001, []common.Hash{hash})
	put("05000f", 0b0000000000000001, 0b0000000000000000, 0b0000000000000001, []common.Hash{hash})
	put("06", 0b0000000000000001, 0b0000000000000000, 0b0000000000000001, []common.Hash{hash})

	integrity.Trie(tx, false, context.Background())

	cursor, err := tx.Cursor(kv.TrieOfAccounts)
	require.NoError(err)
	rl := trie.NewRetainList(0)
	rl.AddHex(common.FromHex("01"))
	rl.AddHex(common.FromHex("0101"))
	rl.AddHex(common.FromHex("030000"))
	rl.AddHex(common.FromHex("03000e"))
	rl.AddHex(common.FromHex("03000e00"))
	rl.AddHex(common.FromHex("0500"))
	var canUse = func(prefix []byte) (bool, []byte) {
		retain, nextCreated := rl.RetainWithMarker(prefix)
		return !retain, nextCreated
	}
	ih := trie.AccTrie(canUse, func(keyHex []byte, _, _, _ uint16, hashes, rootHash []byte) error {
		return nil
	}, cursor, nil)
	k, _, _, _ := ih.AtPrefix([]byte{})
	require.Equal(common.FromHex("0001"), k)
	require.True(ih.SkipState)
	require.Equal([]byte{}, ih.FirstNotCoveredPrefix())
	k, _, _, _ = ih.Next()
	require.Equal(common.FromHex("0100"), k)
	require.True(ih.SkipState)
	require.Equal(common.FromHex("02"), ih.FirstNotCoveredPrefix())
	k, _, _, _ = ih.Next()
	require.Equal(common.FromHex("010100"), k)
	require.True(ih.SkipState)
	k, _, _, _ = ih.Next()
	require.Equal(common.FromHex("010101"), k)
	require.True(ih.SkipState)
	k, _, _, _ = ih.Next()
	require.Equal(common.FromHex("010102"), k)
	require.True(ih.SkipState)
	require.Equal(common.FromHex("1120"), ih.FirstNotCoveredPrefix())
	k, _, _, _ = ih.Next()
	require.Equal(common.FromHex("0102"), k)
	require.True(ih.SkipState)
	require.Equal(common.FromHex("1130"), ih.FirstNotCoveredPrefix())
	k, _, _, _ = ih.Next()
	require.Equal(common.FromHex("020f"), k)
	require.True(ih.SkipState)
	require.Equal(common.FromHex("13"), ih.FirstNotCoveredPrefix())
	k, _, _, _ = ih.Next()
	require.Equal(common.FromHex("03000000"), k)
	require.True(ih.SkipState)
	k, _, _, _ = ih.Next()
	require.Equal(common.FromHex("03000e00000002"), k)
	require.True(ih.SkipState)
	require.Equal(common.FromHex("3001"), ih.FirstNotCoveredPrefix())
	k, _, _, _ = ih.Next()
	require.Equal(common.FromHex("03000e00000e02"), k)
	require.True(ih.SkipState)
	require.Equal(common.FromHex("30e00030"), ih.FirstNotCoveredPrefix())
	k, _, _, _ = ih.Next()
	require.Equal(common.FromHex("05000100"), k)
	require.True(ih.SkipState)
	k, _, _, _ = ih.Next()
	require.Equal(common.FromHex("05000f00"), k)
	require.True(ih.SkipState)
	k, _, _, _ = ih.Next()
	require.Equal(common.FromHex("0600"), k)
	require.True(ih.SkipState)
	k, _, _, _ = ih.Next()
	assert.Nil(t, k)
}
*/
