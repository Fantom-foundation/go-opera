package erigon

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/erigon-lib/kv/memdb"
	"github.com/ledgerwatch/erigon/common"
	"github.com/ledgerwatch/erigon/core/types/accounts"


	"github.com/ethereum/go-ethereum/core/state/snapshot"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/Fantom-foundation/go-opera/gossip/erigon/trie"
	//"github.com/Fantom-foundation/go-opera/gossip/erigon/etrie"


	"github.com/stretchr/testify/assert"

	com "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethdb/memorydb"
)


var (
	key1 = common.HexToHash("0xB1A0000000000000000000000000000000000000000000000000000000000000")
	key2 = []byte("acc-2")
	key3 = []byte("acc-3")
)


func addSnapTestAccount(balance int64) [] byte{
	acc := &snapshot.Account{Balance: big.NewInt(1), Root: emptyRoot.Bytes(), CodeHash: emptyCode.Bytes()}
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
	acc := new(accounts.Account)
	acc.Root = common.Hash(emptyRoot)
	acc.CodeHash = common.Hash(emptyCode)
	acc.Balance.SetUint64(balance)
	
	encoded := make([]byte, acc.EncodingLengthForStorage())
	acc.EncodeForStorage(encoded)

	// hash a key cause trie.NewSecure hashes a key under the hood 
	/*
		func (t *SecureTrie) hashKey(key []byte) []byte {
			h := newHasher(false)
			h.sha.Reset()
			h.sha.Write(key)
			h.sha.Read(t.hashKeyBuf[:])
			returnHasherToPool(h)
			return t.hashKeyBuf[:]
		}
	*/
	
	/*
	hashedKey, err := common.HashData(key)
	if err != nil {
		return nil, err
	}
	*/
	return encoded, tx.Put(kv.HashedAccounts, key1[:], encoded)
}


// legacy and erigon state roots do not match cause they use different serializiation protocols by computing state roots
// legacy implementation applies RLP encoding to serialize trie node and hash serialized data using keccak256 afterwards.
// On an other hand, erigon uses more sophisticated technique in hb.completeLeafHash. THat's why state roots are different.
func TestStateRootsNotMatchWithErigonAccounts(t *testing.T) {
	var (
		diskdb = memorydb.New()
		triedb = trie.NewDatabase(diskdb)

		_, tx = memdb.NewTestTx(t)
	)

	
	
	// 1.make a tree
	//tr, _ := trie.NewSecure(com.Hash{}, triedb)
    tr, err := trie.New(com.Hash{}, triedb)
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

	fmt.Printf("hex value of accountKey: %x\n" , key1.Bytes())
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
	legacyRoot, err := tr.Commit(nil) 
    assert.NoError(t, err)
	//assert.Equal(t, "0xa04693ea110a31037fb5ee814308a6f1d76bdab0b11676bdf4541d2de55ba978", legacyRoot.Hex())



	cfg := StageTrieCfg(nil, false, true, t.TempDir())



	erigonRoot, err := RegenerateIntermediateHashes("IH", tx, cfg, common.Hash{} /* expectedRootHash */, nil /* quit */)
	

	//legacy: "0xe1a85473f43bee6e19dc51a178326327eb61edea2fe1ab6cc5b90c814b1eb371"
	//erigon  : "0x7ed8e10e694f87e13ac1db95f0ebdea4a4644203edcd6b2b9f6c27e31bf1353f"
	assert.Equal(t, legacyRoot.Hex(), erigonRoot.Hex())
}





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




