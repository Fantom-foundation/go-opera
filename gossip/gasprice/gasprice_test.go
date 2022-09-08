package gasprice

import (
	"math/big"
	"testing"

	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"

	"github.com/Fantom-foundation/go-opera/opera"
)

type fakeTx struct {
	gas uint64
	tip *big.Int
	cap *big.Int
}

type TestBackend struct {
	block             idx.Block
	totalGasPowerLeft uint64
	rules             opera.Rules
	pendingRules      opera.Rules
	pendingTxs        []fakeTx
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
	txs := make(types.Transactions, 0, len(t.pendingTxs))
	for _, tx := range t.pendingTxs {
		txs = append(txs, types.NewTx(&types.DynamicFeeTx{
			GasTipCap: tx.tip,
			GasFeeCap: tx.cap,
			Gas:       tx.gas,
		}))
	}
	return txs
}

func TestOracle_EffectiveMinGasPrice(t *testing.T) {
	backend := &TestBackend{
		block:             1,
		totalGasPowerLeft: 0,
		rules:             opera.FakeNetRules(),
		pendingRules:      opera.FakeNetRules(),
	}

	gpo := NewOracle(Config{})
	gpo.cfg.MaxGasPrice = math.MaxBig256
	gpo.cfg.MinGasPrice = new(big.Int)

	// no backend
	require.Equal(t, "0", gpo.EffectiveMinGasPrice().String())
	gpo.backend = backend

	// all the gas is consumed, price should be high
	backend.block++
	backend.totalGasPowerLeft = 0
	require.Equal(t, "25000000000", gpo.EffectiveMinGasPrice().String())

	// test the cache as well
	backend.totalGasPowerLeft = 1008000000
	require.Equal(t, "25000000000", gpo.EffectiveMinGasPrice().String())
	backend.block++
	require.Equal(t, "24994672000", gpo.EffectiveMinGasPrice().String())
	backend.block++

	// all the gas is free, price should be low
	backend.totalGasPowerLeft = gpo.maxTotalGasPower().Uint64()
	require.Equal(t, uint64(0x92aeed1c000), backend.totalGasPowerLeft)
	require.Equal(t, "1000000000", gpo.EffectiveMinGasPrice().String())
	backend.block++

	// edge case with totalGasPowerLeft exceeding maxTotalGasPower
	backend.totalGasPowerLeft = 2 * gpo.maxTotalGasPower().Uint64()
	require.Equal(t, "1000000000", gpo.EffectiveMinGasPrice().String())
	backend.block++

	// half of the gas is free, price should be 3.75x
	backend.totalGasPowerLeft = gpo.maxTotalGasPower().Uint64() / 2
	require.Equal(t, "3750000000", gpo.EffectiveMinGasPrice().String())
	backend.block++

	// third of the gas is free, price should be higher
	backend.totalGasPowerLeft = gpo.maxTotalGasPower().Uint64() / 3
	require.Equal(t, "8125008000", gpo.EffectiveMinGasPrice().String())
	backend.block++

	// check min and max price hard limits don't apply
	gpo.cfg.MaxGasPrice = big.NewInt(2000000000)
	backend.totalGasPowerLeft = gpo.maxTotalGasPower().Uint64() / 3
	require.Equal(t, "8125008000", gpo.EffectiveMinGasPrice().String())
	backend.block++

	gpo.cfg.MinGasPrice = big.NewInt(1500000000)
	backend.totalGasPowerLeft = gpo.maxTotalGasPower().Uint64()
	require.Equal(t, "1000000000", gpo.EffectiveMinGasPrice().String())
	backend.block++
}

func TestOracle_constructiveGasPrice(t *testing.T) {
	backend := &TestBackend{
		totalGasPowerLeft: 0,
		rules:             opera.FakeNetRules(),
		pendingRules:      opera.FakeNetRules(),
	}

	gpo := NewOracle(Config{})
	gpo.backend = backend
	gpo.cfg.MaxGasPrice = math.MaxBig256
	gpo.cfg.MinGasPrice = new(big.Int)

	// all the gas is consumed, price should be high
	backend.totalGasPowerLeft = 0
	require.Equal(t, "2500", gpo.constructiveGasPrice(0, 0, big.NewInt(100)).String())
	require.Equal(t, "2500", gpo.constructiveGasPrice(0, 0.1*DecimalUnit, big.NewInt(100)).String())
	require.Equal(t, "2500", gpo.constructiveGasPrice(1008000000, 0, big.NewInt(100)).String())
	require.Equal(t, "2500", gpo.constructiveGasPrice(gpo.maxTotalGasPower().Uint64()*2, 2*DecimalUnit, big.NewInt(100)).String())

	// all the gas is free, price should be low
	backend.totalGasPowerLeft = gpo.maxTotalGasPower().Uint64()
	require.Equal(t, "100", gpo.constructiveGasPrice(0, 0, big.NewInt(100)).String())
	require.Equal(t, "120", gpo.constructiveGasPrice(0, 0.1*DecimalUnit, big.NewInt(100)).String())
	require.Equal(t, "101", gpo.constructiveGasPrice(100800000000, 0, big.NewInt(100)).String())
	require.Equal(t, "2500", gpo.constructiveGasPrice(gpo.maxTotalGasPower().Uint64()*2, 2*DecimalUnit, big.NewInt(100)).String())

	// half of the gas is free, price should be 3.75x
	backend.totalGasPowerLeft = gpo.maxTotalGasPower().Uint64() / 2
	require.Equal(t, "375", gpo.constructiveGasPrice(0, 0, big.NewInt(100)).String())
	require.Equal(t, "637", gpo.constructiveGasPrice(0, 0.1*DecimalUnit, big.NewInt(100)).String())
	require.Equal(t, "401", gpo.constructiveGasPrice(100800000000, 0, big.NewInt(100)).String())
	require.Equal(t, "2500", gpo.constructiveGasPrice(gpo.maxTotalGasPower().Uint64()*2, 2*DecimalUnit, big.NewInt(100)).String())

	// third of the gas is free, price should be higher
	backend.totalGasPowerLeft = gpo.maxTotalGasPower().Uint64() / 3
	require.Equal(t, "812", gpo.constructiveGasPrice(0, 0, big.NewInt(100)).String())
	require.Equal(t, "1255", gpo.constructiveGasPrice(0, 0.1*DecimalUnit, big.NewInt(100)).String())
	require.Equal(t, "838", gpo.constructiveGasPrice(100800000000, 0, big.NewInt(100)).String())
	require.Equal(t, "2500", gpo.constructiveGasPrice(gpo.maxTotalGasPower().Uint64()*2, 2*DecimalUnit, big.NewInt(100)).String())

}

func TestOracle_reactiveGasPrice(t *testing.T) {
	backend := &TestBackend{
		totalGasPowerLeft: 0,
		rules:             opera.FakeNetRules(),
		pendingRules:      opera.FakeNetRules(),
	}

	gpo := NewOracle(Config{})
	gpo.backend = backend
	gpo.cfg.MaxGasPrice = math.MaxBig256
	gpo.cfg.MinGasPrice = new(big.Int)

	// no stats -> zero price
	gpo.c = circularTxpoolStats{}
	require.Equal(t, "0", gpo.reactiveGasPrice(0).String())
	require.Equal(t, "0", gpo.reactiveGasPrice(DecimalUnit).String())
	gpo.txpoolStatsTick()
	require.Equal(t, "0", gpo.reactiveGasPrice(0).String())
	require.Equal(t, "0", gpo.reactiveGasPrice(DecimalUnit).String())

	// one tx
	gpo.c = circularTxpoolStats{}
	backend.pendingTxs = append(backend.pendingTxs, fakeTx{
		gas: 50000,
		tip: big.NewInt(0),
		cap: big.NewInt(1e9),
	})
	require.Equal(t, "0", gpo.reactiveGasPrice(0).String())
	require.Equal(t, "0", gpo.reactiveGasPrice(DecimalUnit).String())
	gpo.txpoolStatsTick()
	require.Equal(t, "0", gpo.reactiveGasPrice(0).String())
	require.Equal(t, "0", gpo.reactiveGasPrice(0.8*DecimalUnit).String())
	require.Equal(t, "200000000", gpo.reactiveGasPrice(0.9*DecimalUnit).String())
	require.Equal(t, "600000000", gpo.reactiveGasPrice(0.95*DecimalUnit).String())
	require.Equal(t, "920000000", gpo.reactiveGasPrice(0.99*DecimalUnit).String())
	require.Equal(t, "1000000000", gpo.reactiveGasPrice(DecimalUnit).String())

	// add one more tx
	backend.pendingTxs = append(backend.pendingTxs, fakeTx{
		gas: 25000,
		tip: big.NewInt(3 * 1e9),
		cap: big.NewInt(3.5 * 1e9),
	})

	require.Equal(t, "0", gpo.reactiveGasPrice(0).String())
	require.Equal(t, "1000000000", gpo.reactiveGasPrice(DecimalUnit).String())
	gpo.txpoolStatsTick()
	require.Equal(t, "0", gpo.reactiveGasPrice(0).String())
	require.Equal(t, "0", gpo.reactiveGasPrice(0.8*DecimalUnit).String())
	require.Equal(t, "450000000", gpo.reactiveGasPrice(0.9*DecimalUnit).String())
	require.Equal(t, "1350000000", gpo.reactiveGasPrice(0.95*DecimalUnit).String())
	require.Equal(t, "2070000000", gpo.reactiveGasPrice(0.99*DecimalUnit).String())
	require.Equal(t, "2250000000", gpo.reactiveGasPrice(DecimalUnit).String())

	// add two more txs
	backend.pendingTxs = append(backend.pendingTxs, fakeTx{
		gas: 2500000,
		tip: big.NewInt(1 * 1e9),
		cap: big.NewInt(3.5 * 1e9),
	})
	backend.pendingTxs = append(backend.pendingTxs, fakeTx{
		gas: 2500000,
		tip: big.NewInt(0 * 1e9),
		cap: big.NewInt(3.5 * 1e9),
	})

	gpo.txpoolStatsTick()
	require.Equal(t, "0", gpo.reactiveGasPrice(0).String())
	require.Equal(t, "333333333", gpo.reactiveGasPrice(0.8*DecimalUnit).String())
	require.Equal(t, "799999999", gpo.reactiveGasPrice(0.9*DecimalUnit).String())
	require.Equal(t, "1733333332", gpo.reactiveGasPrice(0.95*DecimalUnit).String())
	require.Equal(t, "2479999999", gpo.reactiveGasPrice(0.99*DecimalUnit).String())
	require.Equal(t, "2666666666", gpo.reactiveGasPrice(DecimalUnit).String())
	// price gets closer to latest state
	gpo.txpoolStatsTick()
	require.Equal(t, "500000000", gpo.reactiveGasPrice(0.8*DecimalUnit).String())
	require.Equal(t, "2875000000", gpo.reactiveGasPrice(DecimalUnit).String())
	gpo.txpoolStatsTick()
	require.Equal(t, "600000000", gpo.reactiveGasPrice(0.8*DecimalUnit).String())
	require.Equal(t, "3000000000", gpo.reactiveGasPrice(DecimalUnit).String())
	gpo.txpoolStatsTick()
	require.Equal(t, "666666666", gpo.reactiveGasPrice(0.8*DecimalUnit).String())
	require.Equal(t, "3083333333", gpo.reactiveGasPrice(DecimalUnit).String())
	for i := 0; i < statsBuffer-5; i++ {
		gpo.txpoolStatsTick()
	}
	require.Equal(t, "916666666", gpo.reactiveGasPrice(0.8*DecimalUnit).String())
	require.Equal(t, "3500000000", gpo.reactiveGasPrice(DecimalUnit).String())
	gpo.txpoolStatsTick()
	require.Equal(t, "1000000000", gpo.reactiveGasPrice(0.8*DecimalUnit).String())
	require.Equal(t, "3500000000", gpo.reactiveGasPrice(DecimalUnit).String())
	gpo.txpoolStatsTick()
	require.Equal(t, "1000000000", gpo.reactiveGasPrice(0.8*DecimalUnit).String())
	require.Equal(t, "3500000000", gpo.reactiveGasPrice(DecimalUnit).String())

	// change minGasPrice
	backend.rules.Economy.MinGasPrice = big.NewInt(100)
	gpo.txpoolStatsTick()
	require.Equal(t, "916666675", gpo.reactiveGasPrice(0.8*DecimalUnit).String())
	require.Equal(t, "3458333341", gpo.reactiveGasPrice(DecimalUnit).String())
	gpo.txpoolStatsTick()
	require.Equal(t, "833333350", gpo.reactiveGasPrice(0.8*DecimalUnit).String())
	require.Equal(t, "3416666683", gpo.reactiveGasPrice(DecimalUnit).String())
	for i := 0; i < statsBuffer-3; i++ {
		gpo.txpoolStatsTick()
	}
	require.Equal(t, "83333425", gpo.reactiveGasPrice(0.8*DecimalUnit).String())
	require.Equal(t, "3041666758", gpo.reactiveGasPrice(DecimalUnit).String())
	gpo.txpoolStatsTick()
	require.Equal(t, "100", gpo.reactiveGasPrice(0.8*DecimalUnit).String())
	require.Equal(t, "3000000100", gpo.reactiveGasPrice(DecimalUnit).String())
	gpo.txpoolStatsTick()
	require.Equal(t, "100", gpo.reactiveGasPrice(0.8*DecimalUnit).String())
	require.Equal(t, "3000000100", gpo.reactiveGasPrice(DecimalUnit).String())

	// half of txs are confirmed now
	backend.pendingTxs = backend.pendingTxs[:2]
	gpo.txpoolStatsTick()
	require.Equal(t, "91", gpo.reactiveGasPrice(0.8*DecimalUnit).String())
	require.Equal(t, "3000000100", gpo.reactiveGasPrice(DecimalUnit).String())
	gpo.txpoolStatsTick()
	require.Equal(t, "83", gpo.reactiveGasPrice(0.8*DecimalUnit).String())
	require.Equal(t, "3000000100", gpo.reactiveGasPrice(DecimalUnit).String())
	for i := 0; i < statsBuffer-3; i++ {
		gpo.txpoolStatsTick()
	}
	require.Equal(t, "8", gpo.reactiveGasPrice(0.8*DecimalUnit).String())
	require.Equal(t, "3000000100", gpo.reactiveGasPrice(DecimalUnit).String())
	gpo.txpoolStatsTick()
	require.Equal(t, "0", gpo.reactiveGasPrice(0.8*DecimalUnit).String())
	require.Equal(t, "3000000100", gpo.reactiveGasPrice(DecimalUnit).String())
	gpo.txpoolStatsTick()
	require.Equal(t, "0", gpo.reactiveGasPrice(0.8*DecimalUnit).String())
	require.Equal(t, "3000000100", gpo.reactiveGasPrice(DecimalUnit).String())

	// all txs are confirmed now
	backend.pendingTxs = backend.pendingTxs[:0]
	gpo.txpoolStatsTick()
	require.Equal(t, "0", gpo.reactiveGasPrice(0.8*DecimalUnit).String())
	require.Equal(t, "3000000100", gpo.reactiveGasPrice(DecimalUnit).String())
	gpo.txpoolStatsTick()
	require.Equal(t, "0", gpo.reactiveGasPrice(0.8*DecimalUnit).String())
	require.Equal(t, "3000000100", gpo.reactiveGasPrice(DecimalUnit).String())
	for i := 0; i < statsBuffer-3; i++ {
		gpo.txpoolStatsTick()
	}
	require.Equal(t, "0", gpo.reactiveGasPrice(0.8*DecimalUnit).String())
	require.Equal(t, "3000000100", gpo.reactiveGasPrice(DecimalUnit).String())
	gpo.txpoolStatsTick()
	require.Equal(t, "0", gpo.reactiveGasPrice(0.8*DecimalUnit).String())
	require.Equal(t, "0", gpo.reactiveGasPrice(DecimalUnit).String())
	gpo.txpoolStatsTick()
	require.Equal(t, "0", gpo.reactiveGasPrice(0.8*DecimalUnit).String())
	require.Equal(t, "0", gpo.reactiveGasPrice(DecimalUnit).String())
}
