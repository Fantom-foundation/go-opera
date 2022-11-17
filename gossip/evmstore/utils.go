package evmstore

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/Fantom-foundation/lachesis-base/kvdb/table"
	"github.com/Fantom-foundation/lachesis-base/utils/simplewlru"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/Fantom-foundation/go-opera/utils/iodb"
)

var (
	// EmptyCode is the known hash of the empty EVM bytecode.
	EmptyCode = crypto.Keccak256(nil)

	emptyCodeHash = common.BytesToHash(EmptyCode)
	emptyHash     = common.Hash{}
)

func (s *Store) CheckEvm(forEachState func(func(root common.Hash) (found bool, err error)), onlyRoots bool) error {
	log.Info("Checking every node hash")
	nodeIt := s.table.Evm.NewIterator(nil, nil)
	defer nodeIt.Release()
	for nodeIt.Next() {
		if len(nodeIt.Key()) != 32 {
			continue
		}
		calcHash := crypto.Keccak256(nodeIt.Value())
		if !bytes.Equal(nodeIt.Key(), calcHash) {
			log.Crit("Malformed node record", "exp", common.Bytes2Hex(calcHash), "got", common.Bytes2Hex(nodeIt.Key()))
		}
	}

	log.Info("Checking every code hash")
	codeIt := table.New(s.table.Evm, []byte("c")).NewIterator(nil, nil)
	defer codeIt.Release()
	for codeIt.Next() {
		if len(codeIt.Key()) != 32 {
			continue
		}
		calcHash := crypto.Keccak256(codeIt.Value())
		if !bytes.Equal(codeIt.Key(), calcHash) {
			log.Crit("Malformed code record", "exp", common.Bytes2Hex(calcHash), "got", common.Bytes2Hex(codeIt.Key()))
		}
	}

	log.Info("Checking every preimage")
	preimageIt := table.New(s.table.Evm, []byte("secure-key-")).NewIterator(nil, nil)
	defer preimageIt.Release()
	for preimageIt.Next() {
		if len(preimageIt.Key()) != 32 {
			continue
		}
		calcHash := crypto.Keccak256(preimageIt.Value())
		if !bytes.Equal(preimageIt.Key(), calcHash) {
			log.Crit("Malformed preimage record", "exp", common.Bytes2Hex(calcHash), "got", common.Bytes2Hex(preimageIt.Key()))
		}
	}

	if onlyRoots {
		log.Info("Checking presence of every root")
	} else {
		log.Info("Checking presence of every node")
	}
	var (
		visitedHashes   = make([]common.Hash, 0, 1000000)
		visitedI        = 0
		checkedCache, _ = simplewlru.New(100000000, 100000000)
		cached          = func(h common.Hash) bool {
			_, ok := checkedCache.Get(h)
			return ok
		}
	)
	visited := func(h common.Hash, priority int) {
		base := 100000 * priority
		if visitedI%(1<<(len(visitedHashes)/base)) == 0 {
			visitedHashes = append(visitedHashes, h)
		}
		visitedI++
	}
	forEachState(func(root common.Hash) (found bool, err error) {
		stateTrie, err := s.EvmState.OpenTrie(root)
		found = stateTrie != nil && err == nil
		if !found || onlyRoots {
			return
		}

		// check existence of every code hash and root of every storage trie
		stateIt := stateTrie.NodeIterator(nil)
		for stateItSkip := false; stateIt.Next(!stateItSkip); {
			stateItSkip = false
			if stateIt.Hash() != emptyHash {
				if cached(stateIt.Hash()) {
					stateItSkip = true
					continue
				}
				visited(stateIt.Hash(), 2)
			}

			if stateIt.Leaf() {
				addrHash := common.BytesToHash(stateIt.LeafKey())

				var account state.Account
				if err = rlp.Decode(bytes.NewReader(stateIt.LeafBlob()), &account); err != nil {
					err = fmt.Errorf("Failed to decode accoun as %s addr: %s", addrHash.String(), err.Error())
					return
				}

				codeHash := common.BytesToHash(account.CodeHash)
				if codeHash != emptyCodeHash && !cached(codeHash) {
					code, _ := s.EvmState.ContractCode(addrHash, codeHash)
					if code == nil {
						err = fmt.Errorf("failed to get code %s at %s addr", codeHash.String(), addrHash.String())
						return
					}
					checkedCache.Add(codeHash, true, 1)
				}

				if account.Root != types.EmptyRootHash && !cached(account.Root) {
					storageTrie, storageErr := s.EvmState.OpenStorageTrie(addrHash, account.Root)
					if storageErr != nil {
						err = fmt.Errorf("failed to open storage trie %s at %s addr: %s", account.Root.String(), addrHash.String(), storageErr.Error())
						return
					}
					storageIt := storageTrie.NodeIterator(nil)
					for storageItSkip := false; storageIt.Next(!storageItSkip); {
						storageItSkip = false
						if storageIt.Hash() != emptyHash {
							if cached(storageIt.Hash()) {
								storageItSkip = true
								continue
							}
							visited(storageIt.Hash(), 1)
						}
					}
					if storageIt.Error() != nil {
						err = fmt.Errorf("EVM storage trie %s at %s addr iteration error: %s", account.Root.String(), addrHash.String(), storageIt.Error())
						return
					}
				}
			}
		}

		if stateIt.Error() != nil {
			err = fmt.Errorf("EVM state trie %s iteration error: %s", root.String(), stateIt.Error())
			return
		}
		for _, h := range visitedHashes {
			checkedCache.Add(h, true, 1)
		}
		visitedHashes = visitedHashes[:0]

		return
	})

	return nil
}

func (s *Store) ImportEvm(r io.Reader) error {
	it := iodb.NewIterator(r)
	defer it.Release()
	batch := &restrictedEvmBatch{s.table.Evm.NewBatch()}
	defer batch.Reset()
	for it.Next() {
		err := batch.Put(it.Key(), it.Value())
		if err != nil {
			return err
		}
		if batch.ValueSize() > kvdb.IdealBatchSize {
			err := batch.Write()
			if err != nil {
				return err
			}
			batch.Reset()
		}
	}
	return batch.Write()
}

type restrictedEvmBatch struct {
	kvdb.Batch
}

func IsMptKey(key []byte) bool {
	return len(key) == common.HashLength ||
		(bytes.HasPrefix(key, rawdb.CodePrefix) && len(key) == len(rawdb.CodePrefix)+common.HashLength)
}

func IsPreimageKey(key []byte) bool {
	preimagePrefix := []byte("secure-key-")
	return bytes.HasPrefix(key, preimagePrefix) && len(key) == (len(preimagePrefix)+common.HashLength)
}

func (v *restrictedEvmBatch) Put(key []byte, value []byte) error {
	if !IsMptKey(key) && !IsPreimageKey(key) {
		return errors.New("not expected prefix for EVM history dump")
	}
	return v.Batch.Put(key, value)
}
