package erigon

import (

	"math/big"
	"testing"

	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/erigon-lib/kv/memdb"
	"github.com/ledgerwatch/erigon/common"
	"github.com/ledgerwatch/erigon/core/types/accounts"


	"github.com/ethereum/go-ethereum/core/state/snapshot"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/Fantom-foundation/go-opera/gossip/erigon/trie"


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



func TestCompareEthereumErigonStateRootWithErigonAccounts(t *testing.T) {
	var (
		diskdb = memorydb.New()
		triedb = trie.NewDatabase(diskdb)

		_, tx = memdb.NewTestTx(t)
	)
    // compute state root of 3 test snapshot accounts
	//accMap := make(map[string][]byte)
	
	
	
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

	hashedVal, err := common.HashData(val)
	assert.Nil(t, err)
	t.Log("hashedVal", hashedVal.Hex())



	//replace tr.Update(key1, val) by my implementation below

	// 1. hash a key 
	/*   I replaced this func by common.HashData (used in Erigon)
	func (t *SecureTrie) hashKey(key []byte) []byte {
			h := newHasher(false)
			h.sha.Reset()
			h.sha.Write(key)
			h.sha.Read(t.hashKeyBuf[:])
			returnHasherToPool(h)
			return t.hashKeyBuf[:]
		}
	*/
	

	

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


	// generating erigon state root
	// checkRoot sets to false
	cfg := StageTrieCfg(nil, false, true, t.TempDir())
	//expHash := common.BytesToHash([]byte("ssjj"))

	// without hashing every key "0x82fdcfbe8ec608353eeed139e391c89729101a46c4db43ec6ea1688d6c92125a"
	// with hashing every key "0xa04693ea110a31037fb5ee814308a6f1d76bdab0b11676bdf4541d2de55ba978"
	erigonRoot, err := RegenerateIntermediateHashes("IH", tx, cfg, common.Hash{} /* expectedRootHash */, nil /* quit */)
	
	// legacy and erigon root still do not match
	//expected: "0xe1a85473f43bee6e19dc51a178326327eb61edea2fe1ab6cc5b90c814b1eb371"
	//actual  : "0xa04693ea110a31037fb5ee814308a6f1d76bdab0b11676bdf4541d2de55ba978"
	assert.Equal(t, legacyRoot.Hex(), erigonRoot.Hex())

}

func keybytesToHex(str []byte) []byte {
	l := len(str)*2 + 1
	var nibbles = make([]byte, l)
	for i, b := range str {
		nibbles[i*2] = b / 16
		nibbles[i*2+1] = b % 16
	}
	nibbles[l-1] = 16
	return nibbles
}



