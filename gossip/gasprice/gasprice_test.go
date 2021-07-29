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
	require.Equal(t, "0", gpo.cfg.MinPrice.String())
	require.Equal(t, DefaultMaxPrice.String(), gpo.cfg.MaxPrice.String())
	require.Equal(t, big.NewInt(2*DecimalUnit).String(), gpo.cfg.MiddlePriceMultiplierRatio.String())
	require.Equal(t, big.NewInt(10*DecimalUnit).String(), gpo.cfg.MaxPriceMultiplierRatio.String())
	require.Equal(t, "1", gpo.cfg.GasPowerWallRatio.String())

	gpo = NewOracle(nil, Config{
		GasPowerWallRatio: big.NewInt(2 * DecimalUnit),
	})
	require.Equal(t, "999998", gpo.cfg.GasPowerWallRatio.String())

	gpo = NewOracle(nil, Config{
		MiddlePriceMultiplierRatio: big.NewInt(0.5 * DecimalUnit),
		MaxPriceMultiplierRatio:    big.NewInt(0.5 * DecimalUnit),
	})
	require.Equal(t, DecimalUnitBn.String(), gpo.cfg.MiddlePriceMultiplierRatio.String())
	require.Equal(t, DecimalUnitBn.String(), gpo.cfg.MaxPriceMultiplierRatio.String())

	gpo = NewOracle(nil, Config{
		MiddlePriceMultiplierRatio: big.NewInt(3 * DecimalUnit),
		MaxPriceMultiplierRatio:    big.NewInt(2 * DecimalUnit),
	})
	require.Equal(t, gpo.cfg.MaxPriceMultiplierRatio.String(), gpo.cfg.MiddlePriceMultiplierRatio.String())
}

func TestSuggestPrice(t *testing.T) {
	backend := &TestBackend{
		block:             1,
		totalGasPowerLeft: 0,
		rules:             opera.FakeNetRules(),
		pendingRules:      opera.FakeNetRules(),
	}

	gpo := NewOracle(backend, Config{})

	maxMul := big.NewInt(10 * DecimalUnit)
	gpo.cfg.MiddlePriceMultiplierRatio = big.NewInt(2 * DecimalUnit)
	gpo.cfg.MaxPriceMultiplierRatio = maxMul

	// all the gas is consumed, price should be high
	require.Equal(t, "10000000000", gpo.SuggestPrice().String())

	// increase MaxPriceMultiplierRatio
	maxMul = big.NewInt(100 * DecimalUnit)
	gpo.cfg.MaxPriceMultiplierRatio = maxMul

	// test the cache as well
	require.Equal(t, "10000000000", gpo.SuggestPrice().String())
	backend.block++
	require.Equal(t, "100000000000", gpo.SuggestPrice().String())
	backend.block++

	// all the gas is free, price should be low
	backend.totalGasPowerLeft = gpo.maxTotalGasPower().Uint64()
	require.Equal(t, uint64(0x92aeed1c000), backend.totalGasPowerLeft)
	require.Equal(t, "1000000000", gpo.SuggestPrice().String())
	backend.block++

	// edge case with totalGasPowerLeft exceeding maxTotalGasPower
	backend.totalGasPowerLeft = 2 * gpo.maxTotalGasPower().Uint64()
	require.Equal(t, "1000000000", gpo.SuggestPrice().String())
	backend.block++

	// half of the gas is free, price should be 2x
	backend.totalGasPowerLeft = gpo.maxTotalGasPower().Uint64() / 2
	require.Equal(t, "2000000000", gpo.SuggestPrice().String())
	backend.block++

	// third of the gas is free, price should be higher
	backend.totalGasPowerLeft = gpo.maxTotalGasPower().Uint64() / 3
	require.Equal(t, "34666830000", gpo.SuggestPrice().String())
	backend.block++

	// check the 5% wall
	gpo.cfg.GasPowerWallRatio = big.NewInt(DecimalUnit / 20)
	require.Equal(t, "41543980000", gpo.SuggestPrice().String())
	backend.block++

	// check the 10% wall
	gpo.cfg.GasPowerWallRatio = big.NewInt(DecimalUnit / 10)
	require.Equal(t, "49185334000", gpo.SuggestPrice().String())
	backend.block++

	// check the 20% wall
	gpo.cfg.GasPowerWallRatio = big.NewInt(DecimalUnit / 5)
	require.Equal(t, "67333464000", gpo.SuggestPrice().String())
	backend.block++

	// check the 33.3% wall
	gpo.cfg.GasPowerWallRatio = big.NewInt(DecimalUnit * 0.333)
	require.Equal(t, "99902196000", gpo.SuggestPrice().String())
	backend.block++

	// check the 50.0% wall
	gpo.cfg.GasPowerWallRatio = big.NewInt(DecimalUnit / 2)
	require.Equal(t, "100000000000", gpo.SuggestPrice().String())
	backend.block++

	// check the maximum wall
	gpo.cfg.GasPowerWallRatio = NewOracle(nil, Config{
		GasPowerWallRatio: DecimalUnitBn,
	}).cfg.GasPowerWallRatio
	require.Equal(t, "100000000000", gpo.SuggestPrice().String())
	backend.block++

	// check max price hard limit
	gpo.cfg.MaxPrice = big.NewInt(2000000000)
	backend.totalGasPowerLeft = gpo.maxTotalGasPower().Uint64() / 3
	require.Equal(t, "2000000000", gpo.SuggestPrice().String())
	backend.block++

	// check min price hard limit
	gpo.cfg.MinPrice = big.NewInt(2000000001)
	backend.totalGasPowerLeft = gpo.maxTotalGasPower().Uint64() / 3
	require.Equal(t, "2000000001", gpo.SuggestPrice().String())
	backend.block++
}
