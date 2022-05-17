package gasprice

import (
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/Fantom-foundation/go-opera/opera"
)

type TestBackend struct {
	block             idx.Block
	totalGasPowerLeft uint64
	rules             opera.Rules
	pendingRules      opera.Rules
}

func (t TestBackend) GetLatestBlockIndex() idx.Block {
	return t.block
}

func (t TestBackend) TotalGasPowerLeft() uint64 {
	return t.totalGasPowerLeft
}

func (t TestBackend) GetRules() opera.Rules {
	return t.rules
}

func (t TestBackend) GetPendingRules() opera.Rules {
	return t.pendingRules
}

func (t TestBackend) PendingTxs() types.Transactions {
	return types.Transactions{}
}

// TODO re-implement tests
