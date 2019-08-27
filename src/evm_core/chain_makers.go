package evm_core

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/params"

	"github.com/Fantom-foundation/go-lachesis/src/inter"
)

// BlockGen creates blocks for testing.
// See GenerateChain for a detailed explanation.
type BlockGen struct {
	i       int
	parent  *EvmBlock
	chain   []*EvmBlock
	header  *EvmHeader
	statedb *state.StateDB

	gasPool  *GasPool
	txs      []*types.Transaction
	receipts []*types.Receipt

	config *params.ChainConfig
}

type TestChain struct {
	headers map[common.Hash]*EvmHeader
}

func (tc *TestChain) GetHeader(hash common.Hash, number uint64) *EvmHeader {
	return tc.headers[hash]
}

// SetCoinbase sets the coinbase of the generated block.
// It can be called at most once.
func (b *BlockGen) SetCoinbase(addr common.Address) {
	if b.gasPool != nil {
		if len(b.txs) > 0 {
			panic("coinbase must be set before adding transactions")
		}
		panic("coinbase can only be set once")
	}
	b.header.Coinbase = addr
	b.gasPool = new(GasPool).AddGas(b.header.GasLimit)
}

// AddTx adds a transaction to the generated block. If no coinbase has
// been set, the block's coinbase is set to the zero address.
//
// AddTx panics if the transaction cannot be executed. In addition to
// the protocol-imposed limitations (gas limit, etc.), there are some
// further limitations on the content of transactions that can be
// added. Notably, contract code relying on the BLOCKHASH instruction
// will panic during execution.
func (b *BlockGen) AddTx(tx *types.Transaction) {
	b.AddTxWithChain(nil, tx)
}

// AddTxWithChain adds a transaction to the generated block. If no coinbase has
// been set, the block's coinbase is set to the zero address.
//
// AddTxWithChain panics if the transaction cannot be executed. In addition to
// the protocol-imposed limitations (gas limit, etc.), there are some
// further limitations on the content of transactions that can be
// added. If contract code relies on the BLOCKHASH instruction,
// the block in chain will be returned.
func (b *BlockGen) AddTxWithChain(bc DummyChain, tx *types.Transaction) {
	if b.gasPool == nil {
		b.SetCoinbase(common.Address{})
	}
	b.statedb.Prepare(tx.Hash(), common.Hash{}, len(b.txs))
	receipt, _, _, _, err := ApplyTransaction(b.config, bc, &b.header.Coinbase, b.gasPool, b.statedb, b.header, tx, &b.header.gasUsed, vm.Config{}, false)
	if err != nil {
		panic(err)
	}
	b.txs = append(b.txs, tx)
	b.receipts = append(b.receipts, receipt)
}

// AddUncheckedTx forcefully adds a transaction to the block without any
// validation.
//
// AddUncheckedTx will cause consensus failures when used during real
// chain processing. This is best used in conjunction with raw block insertion.
func (b *BlockGen) AddUncheckedTx(tx *types.Transaction) {
	b.txs = append(b.txs, tx)
}

// Number returns the block number of the block being generated.
func (b *BlockGen) Number() *big.Int {
	return new(big.Int).Set(b.header.Number)
}

// AddUncheckedReceipt forcefully adds a receipts to the block without a
// backing transaction.
//
// AddUncheckedReceipt will cause consensus failures when used during real
// chain processing. This is best used in conjunction with raw block insertion.
func (b *BlockGen) AddUncheckedReceipt(receipt *types.Receipt) {
	b.receipts = append(b.receipts, receipt)
}

// TxNonce returns the next valid transaction nonce for the
// account at addr. It panics if the account does not exist.
func (b *BlockGen) TxNonce(addr common.Address) uint64 {
	if !b.statedb.Exist(addr) {
		panic("account does not exist")
	}
	return b.statedb.GetNonce(addr)
}

// PrevBlock returns a previously generated block by number. It panics if
// num is greater or equal to the number of the block being generated.
// For index -1, PrevBlock returns the parent block given to GenerateChain.
func (b *BlockGen) PrevBlock(index int) *EvmBlock {
	if index >= b.i {
		panic(fmt.Errorf("block index %d out of range (%d,%d)", index, -1, b.i))
	}
	if index == -1 {
		return b.parent
	}
	return b.chain[index]
}

// OffsetTime modifies the time instance of a block, implicitly changing its
// associated difficulty. It's useful to test scenarios where forking is not
// tied to chain length directly.
func (b *BlockGen) OffsetTime(seconds int64) {
	b.header.Time += inter.Timestamp(seconds)
	if b.header.Time <= b.parent.Header().Time {
		panic("block time out of range")
	}
}

// GenerateChain creates a chain of n blocks. The first block's
// parent will be the provided parent. db is used to store
// intermediate states and should contain the parent's state trie.
//
// The generator function is called with a new block generator for
// every block. Any transactions and uncles added to the generator
// become part of the block. If gen is nil, the blocks will be empty
// and their coinbase will be the zero address.
//
// Blocks created by GenerateChain do not contain valid proof of work
// values. Inserting them into BlockChain requires use of FakePow or
// a similar non-validating proof of work implementation.
func GenerateChain(config *params.ChainConfig, parent *EvmBlock, db ethdb.Database, n int, gen func(int, *BlockGen)) ([]*EvmBlock, []types.Receipts, DummyChain) {
	if config == nil {
		config = params.AllEthashProtocolChanges
	}

	chain := &TestChain{
		headers: map[common.Hash]*EvmHeader{},
	}

	blocks, receipts := make([]*EvmBlock, n), make([]types.Receipts, n)
	genblock := func(i int, parent *EvmBlock, statedb *state.StateDB) (*EvmBlock, types.Receipts) {
		b := &BlockGen{i: i, chain: blocks, parent: parent, statedb: statedb, config: config}
		b.header = makeHeader(parent, statedb)

		// Execute any user modifications to the block
		if gen != nil {
			gen(i, b)
		}
		// Finalize and seal the block
		block := &EvmBlock{
			EvmHeader: *b.header,
		}

		// Write state changes to db
		root, err := statedb.Commit(config.IsEIP158(b.header.Number))
		if err != nil {
			panic(fmt.Sprintf("state write error: %v", err))
		}
		if err := statedb.Database().TrieDB().Commit(root, false); err != nil {
			panic(fmt.Sprintf("trie write error: %v", err))
		}
		b.header = block.Header()
		block.Root = root

		return block, b.receipts
	}
	for i := 0; i < n; i++ {
		statedb, err := state.New(parent.Root, state.NewDatabase(db))
		if err != nil {
			panic(err)
		}
		block, receipt := genblock(i, parent, statedb)
		blocks[i] = block
		receipts[i] = receipt
		parent = block

		chain.headers[block.Hash] = block.Header()
	}
	return blocks, receipts, chain
}

func makeHeader(parent *EvmBlock, state *state.StateDB) *EvmHeader {
	var time inter.Timestamp
	if parent.Time == 0 {
		time = 10
	} else {
		time = parent.Time + 10 // block time is fixed at 10 seconds
	}

	return &EvmHeader{
		ParentHash: parent.Hash,
		Coinbase:   parent.Coinbase,
		GasLimit:   parent.GasLimit,
		Number:     new(big.Int).Add(parent.Number, common.Big1),
		Time:       time,
	}
}

// makeHeaderChain creates a deterministic chain of headers rooted at parent.
func makeHeaderChain(parent *EvmHeader, n int, db ethdb.Database, seed int) []*EvmHeader {
	block := &EvmBlock{}
	block.EvmHeader = *parent

	blocks := makeBlockChain(block, n, db, seed)
	headers := make([]*EvmHeader, len(blocks))
	for i, block := range blocks {
		headers[i] = block.Header()
	}
	return headers
}

// makeBlockChain creates a deterministic chain of blocks rooted at parent.
func makeBlockChain(parent *EvmBlock, n int, db ethdb.Database, seed int) []*EvmBlock {
	blocks, _, _ := GenerateChain(params.TestChainConfig, parent, db, n, func(i int, b *BlockGen) {
		b.SetCoinbase(common.Address{0: byte(seed), 19: byte(i)})
	})
	return blocks
}

type fakeChainReader struct {
	config  *params.ChainConfig
	genesis *EvmBlock
}

// Config returns the chain configuration.
func (cr *fakeChainReader) Config() *params.ChainConfig {
	return cr.config
}

func (cr *fakeChainReader) CurrentHeader() *EvmHeader                            { return nil }
func (cr *fakeChainReader) GetHeaderByNumber(number uint64) *EvmHeader           { return nil }
func (cr *fakeChainReader) GetHeaderByHash(hash common.Hash) *EvmHeader          { return nil }
func (cr *fakeChainReader) GetHeader(hash common.Hash, number uint64) *EvmHeader { return nil }
func (cr *fakeChainReader) GetBlock(hash common.Hash, number uint64) *EvmBlock   { return nil }
