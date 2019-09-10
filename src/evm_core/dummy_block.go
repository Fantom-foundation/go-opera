package evm_core

import (
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
)

type (
	EvmHeader struct {
		Number     *big.Int
		Hash       common.Hash
		ParentHash common.Hash
		Root       common.Hash
		Time       inter.Timestamp
		Coinbase   common.Address

		GasLimit uint64
		gasUsed  uint64 // tests only

		Extra []byte
	}

	EvmBlock struct {
		EvmHeader

		Transactions types.Transactions
	}
)

// ToEvmHeader converts inter.Block to EvmHeader.
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

// PrettyHash calcs hash of [genesis] header.
// NOTE: it conflicts with inter.Block.Hash().
// NOTE: it doesn't sum Transactions.
func (b *EvmBlock) PrettyHash() common.Hash {
	e := inter.NewEvent()
	// for nice-looking ID
	e.Epoch = 0
	e.Lamport = idx.Lamport(idx.MaxFrame)
	// actual data hashed
	e.Extra = b.Extra
	e.ClaimedTime = b.Time
	e.TxHash = b.Root
	e.Creator = b.Coinbase

	return common.Hash(e.Hash())
}

func (b *EvmBlock) NumberU64() uint64 {
	return b.Number.Uint64()
}

// Header is a copy of EvmBlock.EvmHeader.
func (b *EvmBlock) Header() *EvmHeader {
	// copy values
	h := b.EvmHeader
	// copy refs
	h.Number = new(big.Int).Set(b.Number)

	return &h
}
