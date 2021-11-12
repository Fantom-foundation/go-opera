package gossip

import (
	"fmt"
	"math/big"
	"sync/atomic"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/state/snapshot"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"

	"github.com/Fantom-foundation/go-opera/evmcore"
)

// ethBlockChain encapsulates functions required to sync a (full or fast) blockchain.
// It is an adapter to replace go-ethereum/core.Blockchain.
type ethBlockChain struct {
	store *Store

	currentBlock     atomic.Value // Current head of the block chain
	currentFastBlock atomic.Value // Current head of the fast-sync chain (may be above the block chain!)
}

func newEthBlockChain(s *Store) (*ethBlockChain, error) {
	bc := &ethBlockChain{
		store: s,
	}

	if err := bc.loadLastState(); err != nil {
		return nil, err
	}
	return bc, nil
}

// loadLastState loads the last known chain state from the database. This method
// assumes that the chain manager mutex is held.
func (bc *ethBlockChain) loadLastState() error {
	header := bc.CurrentHeader()
	block := types.NewBlockWithHeader(header)
	// TODO: add txs

	// Restore the last known heads block
	bc.currentBlock.Store(block)
	bc.currentFastBlock.Store(block)

	return nil
}

// GetTd returns the total difficulty of a local block.
func (bc *ethBlockChain) GetTd(common.Hash, uint64) *big.Int {
	return big.NewInt(0)
}

// StateCache returns the caching database underpinning the blockchain instance.
func (bc *ethBlockChain) StateCache() state.Database {
	return bc.store.EvmStore().EvmDatabase()
}

// ContractCode retrieves a blob of data associated with a contract hash
// either from ephemeral in-memory cache, or from persistent storage.
func (bc *ethBlockChain) ContractCode(hash common.Hash) ([]byte, error) {
	return bc.store.EvmStore().EvmDatabase().ContractCode(common.Hash{}, hash)
}

// HasHeader verifies a header's presence in the local chain.
func (bc *ethBlockChain) HasHeader(h common.Hash, index uint64) bool {
	var empty common.Hash
	if h != empty {
		n := bc.store.GetBlockIndex(hash.Event(h))
		return n != nil
	}

	header := bc.getHeader(idx.Block(index))
	return header != nil
}

// GetHeaderByHash retrieves a header from the local chain.
func (bc *ethBlockChain) GetHeaderByHash(h common.Hash) *types.Header {
	n := bc.store.GetBlockIndex(hash.Event(h))
	if n == nil {
		return nil
	}
	return bc.getHeader(*n)
}

// CurrentHeader retrieves the head header from the local chain.
func (bc *ethBlockChain) CurrentHeader() *types.Header {
	n := bc.store.GetLatestBlockIndex()
	return bc.getHeader(n)
}

// InsertHeaderChain inserts a batch of headers into the local chain.
func (bc *ethBlockChain) InsertHeaderChain([]*types.Header, int) (int, error) {
	panic("ethBlockChain.InsertHeaderChain() call")
	return 0, nil
}

// SetHead rewinds the local chain to a new head.
func (bc *ethBlockChain) SetHead(n uint64) error {
	panic("ethBlockChain.SetHead() call")
	return nil
}

// HasBlock verifies a block's presence in the local chain.
func (bc *ethBlockChain) HasBlock(h common.Hash, index uint64) bool {
	return bc.HasHeader(h, index)
}

// HasFastBlock verifies a fast block's presence in the local chain.
func (bc *ethBlockChain) HasFastBlock(h common.Hash, index uint64) bool {
	return bc.HasHeader(h, index)
}

// GetBlockByHash retrieves a block from the local chain.
func (bc *ethBlockChain) GetBlockByHash(h common.Hash) *types.Block {
	header := bc.GetHeaderByHash(h)
	block := types.NewBlockWithHeader(header)
	// TODO: add txs
	return block
}

// GetBlockByNumber retrieves a block from the database by number, caching it
// (associated with its hash) if found.
func (bc *ethBlockChain) GetBlockByNumber(number uint64) *types.Block {
	header := bc.getHeader(idx.Block(number))
	if header == nil {
		return nil
	}
	return types.NewBlockWithHeader(header)
}

// CurrentBlock retrieves the head block from the local chain.
func (bc *ethBlockChain) CurrentBlock() *types.Block {
	return bc.currentBlock.Load().(*types.Block)
}

// CurrentFastBlock retrieves the head fast block from the local chain.
func (bc *ethBlockChain) CurrentFastBlock() *types.Block {
	return bc.currentFastBlock.Load().(*types.Block)
}

// FastSyncCommitHead directly commits the head block to a certain entity.
func (bc *ethBlockChain) FastSyncCommitHead(h common.Hash) error {
	// Make sure that both the block as well at its state trie exists
	block := bc.GetBlockByHash(h)
	if block == nil {
		return fmt.Errorf("non existent block [%x..]", h[:4])
	}

	// If all checks out, manually set the head block
	bc.currentBlock.Store(block)

	// Destroy any existing state snapshot and regenerate it in the background,
	// also resuming the normal maintenance of any previously paused snapshot.
	bc.store.EvmStore().RebuildEvmSnapshot(block.Root())

	log.Info("Committed new head block", "number", block.Number(), "hash", h)
	return nil
}

// InsertChain inserts a batch of blocks into the local chain.
func (bc *ethBlockChain) InsertChain(types.Blocks) (int, error) {
	panic("ethBlockChain.InsertChain() call")
	return 0, nil
}

// InsertReceiptChain inserts a batch of receipts into the local chain.
func (bc *ethBlockChain) InsertReceiptChain(types.Blocks, []types.Receipts, uint64) (int, error) {
	panic("ethBlockChain.InsertReceiptChain() call")
	return 0, nil
}

// Snapshots returns the blockchain snapshot tree to paused it during sync.
func (bc *ethBlockChain) Snapshots() *snapshot.Tree {
	return bc.store.EvmStore().Snapshots()
}

func (bc *ethBlockChain) getHeader(n idx.Block) *types.Header {
	block := bc.store.GetBlock(n)
	if block == nil {
		return nil
	}

	var prev hash.Event
	if n != 0 {
		prev = bc.store.GetBlock(n - 1).Atropos
	}

	net := bc.store.GetRules() // TODO: get actual rules
	header := evmcore.ToEvmHeader(block, n, prev, net)
	return header.EthHeader()
}

// Stop stops the blockchain service. If any imports are currently in progress
// it will abort them using the procInterrupt.
func (bc *ethBlockChain) Stop() {

}
