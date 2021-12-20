package gossip

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/state/snapshot"
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
func (bc *ethBlockChain) ContractCode(hash common.Hash) ([]byte, error) {
	return bc.store.LastKvdbEvmSnapshot().EvmState.ContractCode(common.Hash{}, hash)
}

// Snapshots returns the blockchain snapshot tree to paused it during sync.
func (bc *ethBlockChain) Snapshots() *snapshot.Tree {
	return bc.store.LastKvdbEvmSnapshot().Snapshots()
}
