package evm_core

import (
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/Fantom-foundation/go-lachesis/inter"
)

type (
	EvmHeader struct {
		Number     *big.Int
		Hash       common.Hash
		ParentHash common.Hash
		Root       common.Hash
		TxHash     common.Hash
		Time       inter.Timestamp
		Coinbase   common.Address

		GasLimit uint64
		GasUsed  uint64
	}

	EvmBlock struct {
		EvmHeader

		Transactions types.Transactions
	}
)

// ToEvmHeader converts inter.Block to EvmHeader.
func ToEvmHeader(block *inter.Block, txHash common.Hash) *EvmHeader {
	return &EvmHeader{
		Hash:       common.Hash(block.Hash()),
		ParentHash: common.Hash(block.PrevHash),
		Root:       common.Hash(block.Root),
		TxHash:     txHash,
		Number:     big.NewInt(int64(block.Index)),
		Time:       block.Time,
		Coinbase:   block.Creator,
		GasLimit:   math.MaxUint64,
		GasUsed:    block.GasUsed,
	}
}

// ConvertFromEthHeader converts ETH-formatted header to Lachesis EVM header
func ConvertFromEthHeader(h *types.Header) *EvmHeader {
	// NOTE: incomplete conversion
	return &EvmHeader{
		Number:     h.Number,
		Coinbase:   h.Coinbase,
		GasLimit:   h.GasLimit,
		GasUsed:    h.GasUsed,
		Root:       h.Root,
		TxHash:     h.TxHash,
		ParentHash: h.ParentHash,
		Time:       inter.FromUnix(int64(h.Time)),
		Hash:       common.BytesToHash(h.Extra),
	}
}

// EthHeader returns header in ETH format
func (h *EvmHeader) EthHeader() *types.Header {
	// NOTE: incomplete conversion
	return &types.Header{
		Number:     h.Number,
		Coinbase:   h.Coinbase,
		GasLimit:   h.GasLimit,
		GasUsed:    h.GasUsed,
		Root:       h.Root,
		TxHash:     h.TxHash,
		ParentHash: h.ParentHash,
		Time:       uint64(h.Time.Unix()),
		Extra:      h.Hash.Bytes(),

		Difficulty: new(big.Int),
	}
}

// Header is a copy of EvmBlock.EvmHeader.
func (b *EvmBlock) Header() *EvmHeader {
	if b == nil {
		return nil
	}
	// copy values
	h := b.EvmHeader
	// copy refs
	h.Number = new(big.Int).Set(b.Number)

	return &h
}

func (b *EvmBlock) NumberU64() uint64 {
	return b.Number.Uint64()
}

func (b *EvmBlock) EthBlock() *types.Block {
	return types.NewBlock(b.EvmHeader.EthHeader(), b.Transactions, nil, nil)
}
