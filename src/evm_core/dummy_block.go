package evm_core

import (
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
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

// HashWith calcs hash of EvmHeader with extra data.
func (header *EvmHeader) HashWith(extra []byte) common.Hash {
	e := inter.NewEvent()
	// for nice-looking ID
	e.Epoch = 0
	e.Lamport = idx.Lamport(idx.MaxFrame)
	// actual data hashed
	e.Extra = extra
	e.ClaimedTime = header.Time
	e.TxHash = header.Root
	e.Creator = header.Coinbase

	return common.Hash(e.Hash())
}

func ToEvmHeader(block *inter.Block) *EvmHeader {
	return &EvmHeader{
		Hash:       common.Hash(block.Hash()),
		ParentHash: common.Hash(block.PrevHash),
		Root:       common.Hash(block.Root),
		Number:     big.NewInt(int64(block.Index)),
		Time:       block.Time,
		Coinbase:   block.Creator,
		GasLimit:   math.MaxUint64,
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
		Root:       b.Root,
		GasLimit:   b.GasLimit,
		gasUsed:    b.gasUsed,
		Number:     new(big.Int).Set(b.Number),
		Time:       b.Time,
		Coinbase:   b.Coinbase,
	}
}
