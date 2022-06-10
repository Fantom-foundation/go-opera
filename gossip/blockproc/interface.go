package blockproc

import (
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"

	"github.com/Fantom-foundation/go-opera/evmcore"
	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/inter/iblockproc"
	"github.com/Fantom-foundation/go-opera/opera"
)

type TxListener interface {
	OnNewLog(*types.Log)
	OnNewReceipt(tx *types.Transaction, r *types.Receipt, originator idx.ValidatorID)
	Finalize() iblockproc.BlockState
	Update(bs iblockproc.BlockState, es iblockproc.EpochState)
}

type TxListenerModule interface {
	Start(block iblockproc.BlockCtx, bs iblockproc.BlockState, es iblockproc.EpochState, statedb *state.StateDB) TxListener
}

type TxTransactor interface {
	PopInternalTxs(block iblockproc.BlockCtx, bs iblockproc.BlockState, es iblockproc.EpochState, sealing bool, statedb *state.StateDB) types.Transactions
}

type SealerProcessor interface {
	EpochSealing() bool
	SealEpoch() (iblockproc.BlockState, iblockproc.EpochState)
	Update(bs iblockproc.BlockState, es iblockproc.EpochState)
}

type SealerModule interface {
	Start(block iblockproc.BlockCtx, bs iblockproc.BlockState, es iblockproc.EpochState) SealerProcessor
}

type ConfirmedEventsProcessor interface {
	ProcessConfirmedEvent(inter.EventI)
	Finalize(block iblockproc.BlockCtx, blockSkipped bool) iblockproc.BlockState
}

type ConfirmedEventsModule interface {
	Start(bs iblockproc.BlockState, es iblockproc.EpochState) ConfirmedEventsProcessor
}

type EVMProcessor interface {
	Execute(txs types.Transactions) types.Receipts
	Finalize() (evmBlock *evmcore.EvmBlock, skippedTxs []uint32, receipts types.Receipts)
}

type EVM interface {
	Start(block iblockproc.BlockCtx, statedb *state.StateDB, reader evmcore.DummyChain, onNewLog func(*types.Log), net opera.Rules, evmCfg *params.ChainConfig) EVMProcessor
}
