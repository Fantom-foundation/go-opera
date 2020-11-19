package blockproc

import (
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/lachesis"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/Fantom-foundation/go-opera/evmcore"
	"github.com/Fantom-foundation/go-opera/inter"
)

type BlockCtx struct {
	Idx    idx.Block
	Time   inter.Timestamp
	CBlock lachesis.Block
}

type TxListener interface {
	OnNewLog(*types.Log)
	OnNewReceipt(tx *types.Transaction, r *types.Receipt, originator idx.ValidatorID)
	Finalize() BlockState
	Update(bs BlockState, es EpochState)
}

type TxListenerModule interface {
	Start(block BlockCtx, bs BlockState, es EpochState, statedb *state.StateDB) TxListener
}

type TxTransactor interface {
	PopInternalTxs(block BlockCtx, bs BlockState, es EpochState, sealing bool, statedb *state.StateDB) types.Transactions
}

type SealerProcessor interface {
	EpochSealing() bool
	SealEpoch() (BlockState, EpochState)
	Update(bs BlockState, es EpochState)
}

type SealerModule interface {
	Start(block BlockCtx, bs BlockState, es EpochState) SealerProcessor
}

type ConfirmedEventsProcessor interface {
	ProcessConfirmedEvent(inter.EventI)
	Finalize(block BlockCtx) BlockState
}

type ConfirmedEventsModule interface {
	Start(bs BlockState, es EpochState) ConfirmedEventsProcessor
}

type EVMProcessor interface {
	Execute(txs types.Transactions, internal bool) types.Receipts
	Finalize() (evmBlock *evmcore.EvmBlock, skippedTxs []uint32, receipts types.Receipts)
}

type EVM interface {
	Start(block BlockCtx, statedb *state.StateDB, reader evmcore.DummyChain, onNewLog func(*types.Log)) EVMProcessor
}
