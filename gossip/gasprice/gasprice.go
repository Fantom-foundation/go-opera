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

	"github.com/Fantom-foundation/go-opera/opera"
	"github.com/Fantom-foundation/go-opera/utils/piecefunc"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
)

var DefaultMaxPrice = big.NewInt(1000000 * params.GWei)

var secondBn = big.NewInt(int64(time.Second))

const DecimalUnit = piecefunc.DecimalUnit

var DecimalUnitBn = big.NewInt(DecimalUnit)

type Config struct {
	MaxPrice                   *big.Int `toml:",omitempty"`
	MinPrice                   *big.Int `toml:",omitempty"`
	MaxPriceMultiplierRatio    *big.Int `toml:",omitempty"`
	MiddlePriceMultiplierRatio *big.Int `toml:",omitempty"`
	GasPowerWallRatio          *big.Int `toml:",omitempty"`
}

type Reader interface {
	GetLatestBlockIndex() idx.Block
	TotalGasPowerLeft() uint64
	GetRules() opera.Rules
	GetPendingRules() opera.Rules
}

// Oracle recommends gas prices based on the content of recent
// blocks. Suitable for both light and full clients.
type Oracle struct {
	backend   Reader
	lastHead  idx.Block
	lastPrice *big.Int

	cfg Config

	cacheLock sync.RWMutex
}

func sanitizeBigInt(val, min, max, _default *big.Int, name string) *big.Int {
	if val == nil || val.Sign() == 0 {
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
	params.MaxPrice = sanitizeBigInt(params.MaxPrice, nil, nil, DefaultMaxPrice, "MaxPrice")
	params.MinPrice = sanitizeBigInt(params.MinPrice, nil, nil, new(big.Int), "MinPrice")
	params.GasPowerWallRatio = sanitizeBigInt(params.GasPowerWallRatio, big.NewInt(1), big.NewInt(DecimalUnit-2), big.NewInt(1), "GasPowerWallRatio")
	params.MaxPriceMultiplierRatio = sanitizeBigInt(params.MaxPriceMultiplierRatio, DecimalUnitBn, nil, big.NewInt(10*DecimalUnit), "MaxPriceMultiplierRatio")
	params.MiddlePriceMultiplierRatio = sanitizeBigInt(params.MiddlePriceMultiplierRatio, DecimalUnitBn, params.MaxPriceMultiplierRatio, big.NewInt(2*DecimalUnit), "MiddlePriceMultiplierRatio")
	return &Oracle{
		backend: backend,
		cfg:     params,
	}
}

func (gpo *Oracle) minGasPrice() *big.Int {
	minPrice := gpo.backend.GetRules().Economy.MinGasPrice
	pendingMinPrice := gpo.backend.GetPendingRules().Economy.MinGasPrice
	if minPrice.Cmp(pendingMinPrice) < 0 {
		minPrice = pendingMinPrice
	}
	if minPrice.Cmp(gpo.cfg.MinPrice) < 0 {
		minPrice = gpo.cfg.MinPrice
	}
	return new(big.Int).Set(minPrice)
}

func (gpo *Oracle) maxTotalGasPower() *big.Int {
	rules := gpo.backend.GetRules()

	allocBn := new(big.Int).SetUint64(rules.Economy.LongGasPower.AllocPerSec)
	periodBn := new(big.Int).SetUint64(uint64(rules.Economy.LongGasPower.MaxAllocPeriod))
	maxTotalGasPowerBn := new(big.Int).Mul(allocBn, periodBn)
	maxTotalGasPowerBn.Div(maxTotalGasPowerBn, secondBn)
	return maxTotalGasPowerBn
}

func (gpo *Oracle) suggestPrice() *big.Int {
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
			Y: gpo.cfg.MaxPriceMultiplierRatio.Uint64(),
		},
		{
			X: gpo.cfg.GasPowerWallRatio.Uint64(),
			Y: gpo.cfg.MaxPriceMultiplierRatio.Uint64(),
		},
		{
			X: gpo.cfg.GasPowerWallRatio.Uint64() + (DecimalUnit-gpo.cfg.GasPowerWallRatio.Uint64())/2,
			Y: gpo.cfg.MiddlePriceMultiplierRatio.Uint64(),
		},
		{
			X: DecimalUnit,
			Y: DecimalUnit,
		},
	})

	multiplier := new(big.Int).SetUint64(multiplierFn(freeRatio))

	// price = multiplier * min gas price
	price := multiplier.Mul(multiplier, gpo.minGasPrice())
	price.Div(price, DecimalUnitBn)
	return price
}

// SuggestPrice returns a gasprice so that newly created transaction can
// have a very high chance to be included in the following blocks.
func (gpo *Oracle) SuggestPrice() *big.Int {
	head := gpo.backend.GetLatestBlockIndex()

	// If the latest gasprice is still available, return it.
	gpo.cacheLock.RLock()
	lastHead, lastPrice := gpo.lastHead, gpo.lastPrice
	gpo.cacheLock.RUnlock()
	if head == lastHead {
		return lastPrice
	}

	price := gpo.suggestPrice()
	if price.Cmp(gpo.cfg.MaxPrice) > 0 {
		price = new(big.Int).Set(gpo.cfg.MaxPrice)
	}
	minimum := gpo.minGasPrice()
	if price.Cmp(minimum) < 0 {
		price = minimum
	}

	gpo.cacheLock.Lock()
	gpo.lastHead = head
	gpo.lastPrice = price
	gpo.cacheLock.Unlock()
	return price
}
