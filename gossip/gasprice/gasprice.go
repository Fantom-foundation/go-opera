// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package gasprice

import (
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/common/math"

	"github.com/Fantom-foundation/go-opera/opera"
	"github.com/Fantom-foundation/go-opera/utils/piecefunc"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
)

var DefaultMaxTipCap = big.NewInt(10000000 * params.GWei)

var secondBn = big.NewInt(int64(time.Second))

const DecimalUnit = piecefunc.DecimalUnit

var DecimalUnitBn = big.NewInt(DecimalUnit)

type Config struct {
	MaxTipCap                   *big.Int `toml:",omitempty"`
	MinTipCap                   *big.Int `toml:",omitempty"`
	MaxTipCapMultiplierRatio    *big.Int `toml:",omitempty"`
	MiddleTipCapMultiplierRatio *big.Int `toml:",omitempty"`
	GasPowerWallRatio           *big.Int `toml:",omitempty"`
}

type Reader interface {
	GetLatestBlockIndex() idx.Block
	TotalGasPowerLeft() uint64
	GetRules() opera.Rules
	GetPendingRules() opera.Rules
}

type cache struct {
	head  idx.Block
	lock  sync.RWMutex
	value *big.Int
}

// Oracle recommends gas prices based on the content of recent
// blocks. Suitable for both light and full clients.
type Oracle struct {
	backend Reader

	cfg Config

	cache cache
}

func sanitizeBigInt(val, min, max, _default *big.Int, name string) *big.Int {
	if val == nil || (val.Sign() == 0 && _default.Sign() != 0) {
		log.Warn(fmt.Sprintf("Sanitizing invalid parameter %s of gasprice oracle", name), "provided", val, "updated", _default)
		return _default
	}
	if min != nil && val.Cmp(min) < 0 {
		log.Warn(fmt.Sprintf("Sanitizing invalid parameter %s of gasprice oracle", name), "provided", val, "updated", min)
		return min
	}
	if max != nil && val.Cmp(max) > 0 {
		log.Warn(fmt.Sprintf("Sanitizing invalid parameter %s of gasprice oracle", name), "provided", val, "updated", max)
		return max
	}
	return val
}

// NewOracle returns a new gasprice oracle which can recommend suitable
// gasprice for newly created transaction.
func NewOracle(backend Reader, params Config) *Oracle {
	params.MaxTipCap = sanitizeBigInt(params.MaxTipCap, nil, nil, DefaultMaxTipCap, "MaxTipCap")
	params.MinTipCap = sanitizeBigInt(params.MinTipCap, nil, nil, new(big.Int), "MinTipCap")
	params.GasPowerWallRatio = sanitizeBigInt(params.GasPowerWallRatio, big.NewInt(1), big.NewInt(DecimalUnit-2), big.NewInt(1), "GasPowerWallRatio")
	params.MaxTipCapMultiplierRatio = sanitizeBigInt(params.MaxTipCapMultiplierRatio, DecimalUnitBn, nil, big.NewInt(10*DecimalUnit), "MaxTipCapMultiplierRatio")
	params.MiddleTipCapMultiplierRatio = sanitizeBigInt(params.MiddleTipCapMultiplierRatio, DecimalUnitBn, params.MaxTipCapMultiplierRatio, big.NewInt(2*DecimalUnit), "MiddleTipCapMultiplierRatio")
	return &Oracle{
		backend: backend,
		cfg:     params,
	}
}

func (gpo *Oracle) maxTotalGasPower() *big.Int {
	rules := gpo.backend.GetRules()

	allocBn := new(big.Int).SetUint64(rules.Economy.LongGasPower.AllocPerSec)
	periodBn := new(big.Int).SetUint64(uint64(rules.Economy.LongGasPower.MaxAllocPeriod))
	maxTotalGasPowerBn := new(big.Int).Mul(allocBn, periodBn)
	maxTotalGasPowerBn.Div(maxTotalGasPowerBn, secondBn)
	return maxTotalGasPowerBn
}

func (gpo *Oracle) suggestTipCap() *big.Int {
	max := gpo.maxTotalGasPower()

	current := new(big.Int).SetUint64(gpo.backend.TotalGasPowerLeft())

	freeRatioBn := current.Mul(current, DecimalUnitBn)
	freeRatioBn.Div(freeRatioBn, max)
	freeRatio := freeRatioBn.Uint64()
	if freeRatio > DecimalUnit {
		freeRatio = DecimalUnit
	}

	multiplierFn := piecefunc.NewFunc([]piecefunc.Dot{
		{
			X: 0,
			Y: gpo.cfg.MaxTipCapMultiplierRatio.Uint64(),
		},
		{
			X: gpo.cfg.GasPowerWallRatio.Uint64(),
			Y: gpo.cfg.MaxTipCapMultiplierRatio.Uint64(),
		},
		{
			X: gpo.cfg.GasPowerWallRatio.Uint64() + (DecimalUnit-gpo.cfg.GasPowerWallRatio.Uint64())/2,
			Y: gpo.cfg.MiddleTipCapMultiplierRatio.Uint64(),
		},
		{
			X: DecimalUnit,
			Y: 0,
		},
	})

	multiplier := new(big.Int).SetUint64(multiplierFn(freeRatio))

	minPrice := gpo.backend.GetRules().Economy.MinGasPrice
	pendingMinPrice := gpo.backend.GetPendingRules().Economy.MinGasPrice
	adjustedMinPrice := math.BigMax(minPrice, pendingMinPrice)

	// tip cap = (multiplier * adjustedMinPrice + adjustedMinPrice) - minPrice
	tip := multiplier.Mul(multiplier, adjustedMinPrice)
	tip.Div(tip, DecimalUnitBn)
	tip.Add(tip, adjustedMinPrice)
	tip.Sub(tip, minPrice)

	if tip.Cmp(gpo.cfg.MinTipCap) < 0 {
		return gpo.cfg.MinTipCap
	}
	if tip.Cmp(gpo.cfg.MaxTipCap) > 0 {
		return gpo.cfg.MaxTipCap
	}
	return tip
}

// SuggestTipCap returns a tip cap so that newly created transaction can have a
// very high chance to be included in the following blocks.
//
// Note, for legacy transactions and the legacy eth_gasPrice RPC call, it will be
// necessary to add the basefee to the returned number to fall back to the legacy
// behavior.
func (gpo *Oracle) SuggestTipCap() *big.Int {
	head := gpo.backend.GetLatestBlockIndex()

	// If the latest gasprice is still available, return it.
	gpo.cache.lock.RLock()
	cachedHead, cachedValue := gpo.cache.head, gpo.cache.value
	gpo.cache.lock.RUnlock()
	if head <= cachedHead {
		return new(big.Int).Set(cachedValue)
	}

	value := gpo.suggestTipCap()

	gpo.cache.lock.Lock()
	if head > gpo.cache.head {
		gpo.cache.head = head
		gpo.cache.value = value
	}
	gpo.cache.lock.Unlock()
	return new(big.Int).Set(value)
}
