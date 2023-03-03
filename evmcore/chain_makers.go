// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package evmcore

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"

	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/opera"
)

// BlockGen creates blocks for testing.
// See GenerateChain for a detailed explanation.
type BlockGen struct {
	i       int
	parent  *EvmBlock
	chain   []*EvmBlock
	header  *EvmHeader
	statedb StateDB

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
	msg, err := TxAsMessage(tx, types.MakeSigner(b.config, b.header.Number), b.header.BaseFee)
	if err != nil {
		panic(err)
	}
	b.statedb.Prepare(tx.Hash(), len(b.txs))
	blockContext := NewEVMBlockContext(b.header, bc, nil)
	vmenv := vm.NewEVM(blockContext, vm.TxContext{}, b.statedb, b.config, opera.DefaultVMConfig)
	receipt, _, _, err := applyTransaction(msg, b.config, b.gasPool, b.statedb, b.header.Number, b.header.Hash, tx, &b.header.GasUsed, vmenv, func(log *types.Log, db StateDB) {})
	if err != nil {
		panic(err)
	}
	b.txs = append(b.txs, tx)
	b.receipts = append(b.receipts, receipt)
}

// GetBalance returns the balance of the given address at the generated block.
func (b *BlockGen) GetBalance(addr common.Address) *big.Int {
	return b.statedb.GetBalance(addr)
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

// BaseFee returns the EIP-1559 base fee of the block being generated.
func (b *BlockGen) BaseFee() *big.Int {
	return new(big.Int).Set(b.header.BaseFee)
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
func GenerateChain(
	config *params.ChainConfig, parent *EvmBlock, n int, stateAt func(common.Hash) StateDB, gen func(int, *BlockGen),
) (
	[]*EvmBlock, []types.Receipts, DummyChain,
) {
	if config == nil {
		config = params.AllEthashProtocolChanges
	}

	chain := &TestChain{
		headers: map[common.Hash]*EvmHeader{},
	}

	blocks, receipts := make([]*EvmBlock, n), make([]types.Receipts, n)
	genblock := func(i int, parent *EvmBlock, statedb StateDB) (*EvmBlock, types.Receipts) {
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
		root, err := flush(statedb, config.IsEIP158(b.header.Number))
		if err != nil {
			panic(fmt.Sprintf("state flush error: %v", err))
		}

		b.header = block.Header()
		block.Root = root

		return block, b.receipts
	}

	for i := 0; i < n; i++ {
		statedb := stateAt(parent.Root)

		block, receipt := genblock(i, parent, statedb)
		blocks[i] = block
		receipts[i] = receipt
		parent = block

		chain.headers[block.Hash] = block.Header()
	}
	return blocks, receipts, chain
}

func makeHeader(parent *EvmBlock, state StateDB) *EvmHeader {
	var t inter.Timestamp
	if parent.Time == 0 {
		t = 10
	} else {
		t = parent.Time + inter.Timestamp(10*time.Second) // block time is fixed at 10 seconds
	}
	header := &EvmHeader{
		ParentHash: parent.Hash,
		Coinbase:   parent.Coinbase,
		GasLimit:   parent.GasLimit,
		BaseFee:    parent.BaseFee,
		Number:     new(big.Int).Add(parent.Number, common.Big1),
		Time:       t,
	}
	return header
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
