package evm_core

import (
	"math/big"

	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type EvmHeader struct {
	Hash       common.Hash
	ParentHash common.Hash

	Root common.Hash

	GasLimit uint64
	gasUsed  uint64 // tests only

	Number *big.Int

	Time inter.Timestamp

	Coinbase common.Address
}

func ToEvmHeader(block *inter.Block) *EvmHeader {
	return &EvmHeader{
		Hash:       common.Hash(block.Hash()),
		ParentHash: common.Hash(block.PrevHash),
		Root:       common.Hash(block.Root),
		Number:     big.NewInt(int64(block.Index)),
		Time:       block.Time,
		Coinbase:   block.Creator,
	}
}

func (b *EvmBlock) NumberU64() uint64 {
	return b.Number.Uint64()
}

type EvmBlock struct {
	EvmHeader

	Transactions types.Transactions
}

func (b *EvmBlock) Header() *EvmHeader {
	return &EvmHeader{
		Hash:       b.Hash,
		ParentHash: b.ParentHash,
		GasLimit:   b.GasLimit,
		gasUsed:    b.gasUsed,
		Number:     new(big.Int).Set(b.Number),
		Time:       b.Time,
		Coinbase:   b.Coinbase,
	}
}
