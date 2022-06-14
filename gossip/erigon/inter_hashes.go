package erigon

import (
	"context"
	"fmt"
	"math/bits"

	"github.com/Fantom-foundation/go-opera/gossip/erigon/etrie"
	"github.com/Fantom-foundation/go-opera/gossip/erigon/benchtrie"
	
	"github.com/ledgerwatch/erigon-lib/common/length"
	"github.com/ledgerwatch/erigon-lib/etl"
	"github.com/ledgerwatch/erigon-lib/kv"

	"github.com/ledgerwatch/erigon/cmd/rpcdaemon/interfaces"
	"github.com/ledgerwatch/erigon/common"


	"github.com/ledgerwatch/log/v3"
)

type TrieCfg struct {
	db                kv.RwDB
	checkRoot         bool
	tmpDir            string
	saveNewHashesToDB bool // no reason to save changes when calculating root for mining
	blockReader       interfaces.FullBlockReader
}

func StageTrieCfg(db kv.RwDB, checkRoot, saveNewHashesToDB bool, tmpDir string) TrieCfg {
	return TrieCfg{
		db:                db,
		checkRoot:         checkRoot,
		saveNewHashesToDB: saveNewHashesToDB,
		tmpDir:            tmpDir,
	}
}

func assertSubset(a, b uint16) {
	if (a & b) != a { // a & b == a - checks whether a is subset of b
		panic(fmt.Errorf("invariant 'is subset' failed: %b, %b", a, b))
	}
}

func accountTrieCollector(collector *etl.Collector) benchtrie.HashCollector2 {
	newV := make([]byte, 0, 1024)
	return func(keyHex []byte, hasState, hasTree, hasHash uint16, hashes, _ []byte) error {
		if len(keyHex) == 0 {
			return nil
		}
		if hasState == 0 {
			return collector.Collect(keyHex, nil)
		}
		if bits.OnesCount16(hasHash) != len(hashes)/length.Hash {
			panic(fmt.Errorf("invariant bits.OnesCount16(hasHash) == len(hashes) failed: %d, %d", bits.OnesCount16(hasHash), len(hashes)/length.Hash))
		}
		assertSubset(hasTree, hasState)
		assertSubset(hasHash, hasState)
		newV = etrie.MarshalTrieNode(hasState, hasTree, hasHash, hashes, nil, newV)
		return collector.Collect(keyHex, newV)
	}
}

func storageTrieCollector(collector *etl.Collector) benchtrie.StorageHashCollector2 {
	newK := make([]byte, 0, 128)
	newV := make([]byte, 0, 1024)
	return func(accWithInc []byte, keyHex []byte, hasState, hasTree, hasHash uint16, hashes, rootHash []byte) error {
		newK = append(append(newK[:0], accWithInc...), keyHex...)
		if hasState == 0 {
			return collector.Collect(newK, nil)
		}
		if len(keyHex) > 0 && hasHash == 0 && hasTree == 0 {
			return nil
		}
		if bits.OnesCount16(hasHash) != len(hashes)/length.Hash {
			panic(fmt.Errorf("invariant bits.OnesCount16(hasHash) == len(hashes) failed: %d, %d", bits.OnesCount16(hasHash), len(hashes)/length.Hash))
		}
		assertSubset(hasTree, hasState)
		assertSubset(hasHash, hasState)
		newV = etrie.MarshalTrieNode(hasState, hasTree, hasHash, hashes, rootHash, newV)
		return collector.Collect(newK, newV)
	}
}

func accountTrieCollectorBench(collector *etl.Collector) benchtrie.HashCollector2 {
	newV := make([]byte, 0, 1024)
	return func(keyHex []byte, hasState, hasTree, hasHash uint16, hashes, _ []byte) error {
		if len(keyHex) == 0 {
			return nil
		}
		if hasState == 0 {
			return collector.Collect(keyHex, nil)
		}
		if bits.OnesCount16(hasHash) != len(hashes)/length.Hash {
			panic(fmt.Errorf("invariant bits.OnesCount16(hasHash) == len(hashes) failed: %d, %d", bits.OnesCount16(hasHash), len(hashes)/length.Hash))
		}
		assertSubset(hasTree, hasState)
		assertSubset(hasHash, hasState)
		newV = etrie.MarshalTrieNode(hasState, hasTree, hasHash, hashes, nil, newV)
		return collector.Collect(keyHex, newV)
	}
}

func storageTrieCollectorBench(collector *etl.Collector) benchtrie.StorageHashCollector2 {
	newK := make([]byte, 0, 128)
	newV := make([]byte, 0, 1024)
	return func(accWithInc []byte, keyHex []byte, hasState, hasTree, hasHash uint16, hashes, rootHash []byte) error {
		newK = append(append(newK[:0], accWithInc...), keyHex...)
		if hasState == 0 {
			return collector.Collect(newK, nil)
		}
		if len(keyHex) > 0 && hasHash == 0 && hasTree == 0 {
			return nil
		}
		if bits.OnesCount16(hasHash) != len(hashes)/length.Hash {
			panic(fmt.Errorf("invariant bits.OnesCount16(hasHash) == len(hashes) failed: %d, %d", bits.OnesCount16(hasHash), len(hashes)/length.Hash))
		}
		assertSubset(hasTree, hasState)
		assertSubset(hasHash, hasState)
		newV = etrie.MarshalTrieNode(hasState, hasTree, hasHash, hashes, rootHash, newV)
		return collector.Collect(newK, newV)
	}
}

// refactored RegenerateIntermediateHashes
func ComputeStateRoot(logPrefix string, db kv.RwDB, cfg TrieCfg,

//expectedRootHash common.Hash
) (common.Hash, error) {
	log.Info(fmt.Sprintf("[%s] Generation of trie hashes started", logPrefix))
	defer log.Info(fmt.Sprintf("[%s] Generation ended", logPrefix))

	tx, err := db.BeginRw(context.Background())
	if err != nil {
		return etrie.EmptyRoot, err
	}
	_ = tx.ClearBucket(kv.TrieOfAccounts)
	_ = tx.ClearBucket(kv.TrieOfStorage)

	accTrieCollector := etl.NewCollector(logPrefix, cfg.tmpDir,
		etl.NewSortableBuffer(etl.BufferOptimalSize))
	defer accTrieCollector.Close()
	accTrieCollectorFunc := accountTrieCollector(accTrieCollector)

	stTrieCollector := etl.NewCollector(logPrefix, cfg.tmpDir,
		etl.NewSortableBuffer(etl.BufferOptimalSize))
	defer stTrieCollector.Close()
	stTrieCollectorFunc := storageTrieCollector(stTrieCollector)

	loader := benchtrie.NewFlatDBTrieLoader(logPrefix)
	if err := loader.Reset(etrie.NewRetainList(0), accTrieCollectorFunc,
		stTrieCollectorFunc, false); err != nil {
		return etrie.EmptyRoot, err
	}
	hash, err := loader.CalcTrieRoot(tx, []byte{}, nil)
	if err != nil {
		return etrie.EmptyRoot, err
	}

	/*
		if cfg.checkRoot && hash != expectedRootHash {
			return hash, nil
		}
	*/
	//log.Info(fmt.Sprintf("[%s] Trie root", logPrefix), "hash", hash.Hex())

	if err := accTrieCollector.Load(tx, kv.TrieOfAccounts, etl.IdentityLoadFunc,
		etl.TransformArgs{Quit: nil}); err != nil {
		return etrie.EmptyRoot, err
	}
	if err := stTrieCollector.Load(tx, kv.TrieOfStorage, etl.IdentityLoadFunc,
		etl.TransformArgs{Quit: nil}); err != nil {
		return etrie.EmptyRoot, err
	}

	// ?
	if err := tx.Commit(); err != nil {
		return etrie.EmptyRoot, err
	}
	return hash, nil

}

// for debugging purposes, TODO make it private 
/*
func RegenerateIntermediateHashes(logPrefix string,
	db kv.RwTx, cfg TrieCfg,
	expectedRootHash common.Hash,
	quit <-chan struct{}) (common.Hash, error) {
	log.Info(fmt.Sprintf("[%s] Regeneration trie hashes started", logPrefix))
	defer log.Info(fmt.Sprintf("[%s] Regeneration ended", logPrefix))
	_ = db.ClearBucket(kv.TrieOfAccounts)
	_ = db.ClearBucket(kv.TrieOfStorage)

	accTrieCollector := etl.NewCollector(logPrefix, cfg.tmpDir,
		etl.NewSortableBuffer(etl.BufferOptimalSize))
	defer accTrieCollector.Close()
	accTrieCollectorFunc := accountTrieCollector(accTrieCollector)

	stTrieCollector := etl.NewCollector(logPrefix, cfg.tmpDir,
		etl.NewSortableBuffer(etl.BufferOptimalSize))
	defer stTrieCollector.Close()
	stTrieCollectorFunc := storageTrieCollector(stTrieCollector)
    
	loader := etrie.NewFlatDBTrieLoader(logPrefix)
	if err := loader.Reset(etrie.NewRetainList(0), accTrieCollectorFunc,
		stTrieCollectorFunc, false); err != nil {
		return etrie.EmptyRoot, err
	}
	hash, err := loader.CalcTrieRoot(db, []byte{}, quit)
	if err != nil {
		return etrie.EmptyRoot, err
	}

	log.Info("CalcTrieRoot Hash", hash.Hex())

	if cfg.checkRoot && hash != expectedRootHash {
		return hash, fmt.Errorf("Calculated Hash %s does not match expectedRootHash %s ", hash.Hex(), expectedRootHash.Hex())
	}
	log.Info(fmt.Sprintf("[%s] Trie root", logPrefix), "hash", hash.Hex())

	if err := accTrieCollector.Load(db, kv.TrieOfAccounts, etl.IdentityLoadFunc,
		etl.TransformArgs{Quit: quit}); err != nil {
		return etrie.EmptyRoot, err
	}
	if err := stTrieCollector.Load(db, kv.TrieOfStorage, etl.IdentityLoadFunc,
		etl.TransformArgs{Quit: quit}); err != nil {
		return etrie.EmptyRoot, err
	}
	return hash, nil
}
*/

// for bench purposes
func RegenerateIntermediateHashesBench(logPrefix string,
	db kv.RwTx, cfg TrieCfg,
	expectedRootHash common.Hash,
	quit <-chan struct{}) (common.Hash, error) {
	log.Info(fmt.Sprintf("[%s] Regeneration trie hashes started", logPrefix))
	defer log.Info(fmt.Sprintf("[%s] Regeneration ended", logPrefix))
	_ = db.ClearBucket(kv.TrieOfAccounts)
	_ = db.ClearBucket(kv.TrieOfStorage)

	accTrieCollector := etl.NewCollector(logPrefix, cfg.tmpDir,
		etl.NewSortableBuffer(etl.BufferOptimalSize))
	defer accTrieCollector.Close()
	accTrieCollectorFunc := accountTrieCollectorBench(accTrieCollector)

	stTrieCollector := etl.NewCollector(logPrefix, cfg.tmpDir,
		etl.NewSortableBuffer(etl.BufferOptimalSize))
	defer stTrieCollector.Close()
	stTrieCollectorFunc := storageTrieCollectorBench(stTrieCollector)
    
	loader := benchtrie.NewFlatDBTrieLoader(logPrefix)
	if err := loader.Reset(benchtrie.NewRetainList(0), accTrieCollectorFunc,
		stTrieCollectorFunc, false); err != nil {
		return etrie.EmptyRoot, err
	}
	hash, err := loader.CalcTrieRoot(db, []byte{}, quit)
	if err != nil {
		return etrie.EmptyRoot, err
	}

	log.Info("CalcTrieRoot Hash", hash.Hex())

	if cfg.checkRoot && hash != expectedRootHash {
		return hash, fmt.Errorf("Calculated Hash %s does not match expectedRootHash %s ", hash.Hex(), expectedRootHash.Hex())
	}
	log.Info(fmt.Sprintf("[%s] Trie root", logPrefix), "hash", hash.Hex())

	if err := accTrieCollector.Load(db, kv.TrieOfAccounts, etl.IdentityLoadFunc,
		etl.TransformArgs{Quit: quit}); err != nil {
		return benchtrie.EmptyRoot, err
	}
	if err := stTrieCollector.Load(db, kv.TrieOfStorage, etl.IdentityLoadFunc,
		etl.TransformArgs{Quit: quit}); err != nil {
		return benchtrie.EmptyRoot, err
	}
	return hash, nil
}
