package gasprice

import (
	"math/big"
	"testing"

	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/stretchr/testify/require"

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

func TestConstructor(t *testing.T) {
	gpo := NewOracle(nil, Config{})
	require.Equal(t, "0", gpo.cfg.MinTipCap.String())
	require.Equal(t, DefaultMaxTipCap.String(), gpo.cfg.MaxTipCap.String())
	require.Equal(t, big.NewInt(2*DecimalUnit).String(), gpo.cfg.MiddleTipCapMultiplierRatio.String())
	require.Equal(t, big.NewInt(10*DecimalUnit).String(), gpo.cfg.MaxTipCapMultiplierRatio.String())
	require.Equal(t, "1", gpo.cfg.GasPowerWallRatio.String())

	gpo = NewOracle(nil, Config{
		GasPowerWallRatio: big.NewInt(2 * DecimalUnit),
	})
	require.Equal(t, "999998", gpo.cfg.GasPowerWallRatio.String())

	gpo = NewOracle(nil, Config{
		MiddleTipCapMultiplierRatio: big.NewInt(0.5 * DecimalUnit),
		MaxTipCapMultiplierRatio:    big.NewInt(0.5 * DecimalUnit),
	})
	require.Equal(t, DecimalUnitBn.String(), gpo.cfg.MiddleTipCapMultiplierRatio.String())
	require.Equal(t, DecimalUnitBn.String(), gpo.cfg.MaxTipCapMultiplierRatio.String())

	gpo = NewOracle(nil, Config{
		MiddleTipCapMultiplierRatio: big.NewInt(3 * DecimalUnit),
		MaxTipCapMultiplierRatio:    big.NewInt(2 * DecimalUnit),
	})
	require.Equal(t, gpo.cfg.MaxTipCapMultiplierRatio.String(), gpo.cfg.MiddleTipCapMultiplierRatio.String())
}

func TestSuggestTipCap(t *testing.T) {
	backend := &TestBackend{
		block:             1,
		totalGasPowerLeft: 0,
		rules:             opera.FakeNetRules(),
		pendingRules:      opera.FakeNetRules(),
	}

	gpo := NewOracle(backend, Config{})

	maxMul := big.NewInt(9 * DecimalUnit)
	gpo.cfg.MiddleTipCapMultiplierRatio = big.NewInt(DecimalUnit)
	gpo.cfg.MaxTipCapMultiplierRatio = maxMul

	// all the gas is consumed, price should be high
	require.Equal(t, "9000000000", gpo.SuggestTipCap().String())

	// increase MaxTipCapMultiplierRatio
	maxMul = big.NewInt(100 * DecimalUnit)
	gpo.cfg.MaxTipCapMultiplierRatio = maxMul

	// test the cache as well
	require.Equal(t, "9000000000", gpo.SuggestTipCap().String())
	backend.block++
	require.Equal(t, "100000000000", gpo.SuggestTipCap().String())
	backend.block++

	// all the gas is free, price should be low
	backend.totalGasPowerLeft = gpo.maxTotalGasPower().Uint64()
	require.Equal(t, uint64(0x92aeed1c000), backend.totalGasPowerLeft)
	require.Equal(t, "0", gpo.SuggestTipCap().String())
	backend.block++

	// edge case with totalGasPowerLeft exceeding maxTotalGasPower
	backend.totalGasPowerLeft = 2 * gpo.maxTotalGasPower().Uint64()
	require.Equal(t, "0", gpo.SuggestTipCap().String())
	backend.block++

	// half of the gas is free, price should be 2x
	backend.totalGasPowerLeft = gpo.maxTotalGasPower().Uint64() / 2
	require.Equal(t, "1000000000", gpo.SuggestTipCap().String())
	backend.block++

	// third of the gas is free, price should be higher
	backend.totalGasPowerLeft = gpo.maxTotalGasPower().Uint64() / 3
	require.Equal(t, "34000165000", gpo.SuggestTipCap().String())
	backend.block++

	// check the 5% wall
	gpo.cfg.GasPowerWallRatio = big.NewInt(DecimalUnit / 20)
	require.Equal(t, "40947490000", gpo.SuggestTipCap().String())
	backend.block++

	// check the 10% wall
	gpo.cfg.GasPowerWallRatio = big.NewInt(DecimalUnit / 10)
	require.Equal(t, "48666817000", gpo.SuggestTipCap().String())
	backend.block++

	// check the 20% wall
	gpo.cfg.GasPowerWallRatio = big.NewInt(DecimalUnit / 5)
	require.Equal(t, "67000132000", gpo.SuggestTipCap().String())
	backend.block++

	// check the 33.3% wall
	gpo.cfg.GasPowerWallRatio = big.NewInt(DecimalUnit * 0.333)
	require.Equal(t, "99901198000", gpo.SuggestTipCap().String())
	backend.block++

	// check the 50.0% wall
	gpo.cfg.GasPowerWallRatio = big.NewInt(DecimalUnit / 2)
	require.Equal(t, "100000000000", gpo.SuggestTipCap().String())
	backend.block++

	// check the maximum wall
	gpo.cfg.GasPowerWallRatio = NewOracle(nil, Config{
		GasPowerWallRatio: DecimalUnitBn,
	}).cfg.GasPowerWallRatio
	require.Equal(t, "100000000000", gpo.SuggestTipCap().String())
	backend.block++

	// check max price hard limit
	gpo.cfg.MaxTipCap = big.NewInt(2000000000)
	backend.totalGasPowerLeft = gpo.maxTotalGasPower().Uint64() / 3
	require.Equal(t, "2000000000", gpo.SuggestTipCap().String())
	backend.block++

	// check min price hard limit
	gpo.cfg.MinTipCap = big.NewInt(1500000000)
	backend.totalGasPowerLeft = gpo.maxTotalGasPower().Uint64()
	require.Equal(t, "1500000000", gpo.SuggestTipCap().String())
	backend.block++
}
