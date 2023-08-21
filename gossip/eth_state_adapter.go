package gossip

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/state/snapshot"
	"github.com/ethereum/go-ethereum/trie"
)

// ethBlockChain wraps store to implement eth/protocols/snap.BlockChain interface.
type ethBlockChain struct {
	store *Store
}

func newEthBlockChain(s *Store) (*ethBlockChain, error) {
	bc := &ethBlockChain{
		store: s,
	}

	return bc, nil
}

// StateCache returns the caching database underpinning the blockchain instance.
func (bc *ethBlockChain) StateCache() state.Database {
	return bc.store.LastKvdbEvmSnapshot().EvmState
}

// ContractCode retrieves a blob of data associated with a contract hash
// either from ephemeral in-memory cache, or from persistent storage.
func (bc *ethBlockChain) ContractCodeWithPrefix(hash common.Hash) ([]byte, error) {
	return bc.store.LastKvdbEvmSnapshot().EvmState.ContractCode(common.Address{}, hash)
}

// Snapshots returns the blockchain snapshot tree to paused it during sync.
func (bc *ethBlockChain) Snapshots() *snapshot.Tree {
	return bc.store.LastKvdbEvmSnapshot().Snapshots()
}

// TrieDB retrieves the low level trie database used for data storage.
func (bc *ethBlockChain) TrieDB() *trie.Database {
	return bc.store.evm.EvmState.TrieDB()
}
