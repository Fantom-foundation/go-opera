package erigon

import (
	"encoding/binary"
	"testing"

	"github.com/ledgerwatch/erigon-lib/common/length"
	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/erigon-lib/kv/memdb"
	"github.com/ledgerwatch/erigon/common"
	"github.com/ledgerwatch/erigon/common/dbutils"
	"github.com/ledgerwatch/erigon/core/types/accounts"
	"github.com/ledgerwatch/erigon/params"
	//"github.com/ledgerwatch/erigon/turbo/snapshotsync"
	"github.com/ledgerwatch/erigon/turbo/trie"

	"github.com/stretchr/testify/assert"
)


func addTestAccount(tx kv.Putter, hash common.Hash, balance uint64, incarnation uint64) error {
	acc := accounts.NewAccount()
	acc.Balance.SetUint64(balance)
	acc.Incarnation = incarnation
	if incarnation != 0 {
		acc.CodeHash = common.HexToHash("0x5be74cad16203c4905c068b012a2e9fb6d19d036c410f16fd177f337541440dd")
	}
	encoded := make([]byte, acc.EncodingLengthForStorage())
	acc.EncodeForStorage(encoded)
	return tx.Put(kv.HashedAccounts, hash[:], encoded)
}

func TestAccountAndStorageTrie(t *testing.T) {
	_, tx := memdb.NewTestTx(t)

	hash1 := common.HexToHash("0xB000000000000000000000000000000000000000000000000000000000000000")
	assert.Nil(t, addTestAccount(tx, hash1, 3*params.Ether, 0))

	hash2 := common.HexToHash("0xB040000000000000000000000000000000000000000000000000000000000000")
	assert.Nil(t, addTestAccount(tx, hash2, 1*params.Ether, 0))

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

	// ----------------------------------------------------------------
	// Populate account & storage trie DB tables
	// ----------------------------------------------------------------

	cfg := StageTrieCfg(nil, false, true, t.TempDir())
	expHash := common.BytesToHash([]byte("ssjj"))
	hash, err := RegenerateIntermediateHashes("IH", tx, cfg, expHash /* expectedRootHash */, nil /* quit */)
	assert.NotNil(t, hash.Hex())
	assert.Nil(t, err)

	// ----------------------------------------------------------------
	// Check account trie
	// ----------------------------------------------------------------

	accountTrieA := make(map[string][]byte)
	err = tx.ForEach(kv.TrieOfAccounts, nil, func(k, v []byte) error {
		accountTrieA[string(k)] = common.CopyBytes(v)
		return nil
	})
	assert.Nil(t, err)

	assert.Equal(t, 2, len(accountTrieA))

	hasState1a, hasTree1a, hasHash1a, hashes1a, rootHash1a := trie.UnmarshalTrieNode(accountTrieA[string(common.FromHex("0B"))])
	assert.Equal(t, uint16(0b1011), hasState1a)
	assert.Equal(t, uint16(0b0001), hasTree1a)
	assert.Equal(t, uint16(0b1001), hasHash1a)
	assert.Equal(t, 2*length.Hash, len(hashes1a))
	assert.Equal(t, 0, len(rootHash1a))

	hasState2a, hasTree2a, hasHash2a, hashes2a, rootHash2a := trie.UnmarshalTrieNode(accountTrieA[string(common.FromHex("0B00"))])
	assert.Equal(t, uint16(0b10001), hasState2a)
	assert.Equal(t, uint16(0b00000), hasTree2a)
	assert.Equal(t, uint16(0b10000), hasHash2a)
	assert.Equal(t, 1*length.Hash, len(hashes2a))
	assert.Equal(t, 0, len(rootHash2a))

	// ----------------------------------------------------------------
	// Check storage trie
	// ----------------------------------------------------------------

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

	hasState3, hasTree3, hasHash3, hashes3, rootHash3 := trie.UnmarshalTrieNode(storageTrie[string(storageKey)])
	assert.Equal(t, uint16(0b1010), hasState3)
	assert.Equal(t, uint16(0b0000), hasTree3)
	assert.Equal(t, uint16(0b0010), hasHash3)
	assert.Equal(t, 1*length.Hash, len(hashes3))
	assert.Equal(t, length.Hash, len(rootHash3))
}

