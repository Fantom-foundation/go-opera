package gasprice

import (
	"math/big"
	"sort"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/core/types"

	"github.com/Fantom-foundation/go-opera/utils/piecefunc"
)

const (
	percentilesPerStat = 40
	statUpdatePeriod   = time.Second
	statsBuffer        = (8 * time.Second) / statUpdatePeriod // 10
)

type txpoolStat struct {
	totalGas    uint64
	percentiles [percentilesPerStat]*big.Int
}

type circularTxpoolStats struct {
	stats     [statsBuffer]txpoolStat
	i         int
	activated uint32
	avg       atomic.Value
}

var certaintyToTtf = piecefunc.NewFunc([]piecefunc.Dot{
	{
		X: 0,
		Y: 10000 * uint64(time.Millisecond),
	},
	{
		X: 0.2 * DecimalUnit,
		Y: 4000 * uint64(time.Millisecond),
	},
	{
		X: 0.5 * DecimalUnit,
		Y: 1500 * uint64(time.Millisecond),
	},
	{
		X: DecimalUnit,
		Y: 0,
	},
})

func (gpo *Oracle) reactiveGasPrice(certainty uint64) *big.Int {
	allocPerSec := gpo.backend.GetRules().Economy.LongGasPower.AllocPerSec
	targetTtf := certaintyToTtf(certainty)

	ttfAllocGasBn := new(big.Int).Mul(new(big.Int).SetUint64(allocPerSec), new(big.Int).SetUint64(targetTtf))
	ttfAllocGasBn.Div(ttfAllocGasBn, secondBn)
	return gpo.c.getGasPriceForGasAbove(ttfAllocGasBn.Uint64())
}

func (gpo *Oracle) txpoolStatsTick() {
	c := &gpo.c
	// calculate txpool statistic and push into the circular buffer
	c.stats[c.i] = gpo.calcTxpoolStat()
	c.i = (c.i + 1) % len(c.stats)
	// calculate average of statistics in the circular buffer
	c.avg.Store(c.calcAvg())
}

func (gpo *Oracle) txpoolStatsLoop() {
	ticker := time.NewTicker(statUpdatePeriod)
	defer ticker.Stop()
	for i := uint32(0); ; i++ {
		select {
		case <-ticker.C:
			// calculate more frequently after first request
			if atomic.LoadUint32(&gpo.c.activated) != 0 || i%5 == 0 {
				gpo.txpoolStatsTick()
			}
		case <-gpo.quit:
			return
		}
	}
}

// calcAvg calculates average of statistics in the circular buffer
func (c *circularTxpoolStats) calcAvg() txpoolStat {
	avg := txpoolStat{}
	for p := range avg.percentiles {
		avg.percentiles[p] = new(big.Int)
	}
	nonZero := uint64(0)
	for _, s := range c.stats {
		if s.totalGas == 0 {
			continue
		}
		nonZero++
		avg.totalGas += s.totalGas
		for p := range s.percentiles {
			avg.percentiles[p].Add(avg.percentiles[p], s.percentiles[p])
		}
	}
	if nonZero == 0 {
		return avg
	}
	avg.totalGas /= nonZero
	nonZeroBn := new(big.Int).SetUint64(nonZero)
	for p := range avg.percentiles {
		avg.percentiles[p].Div(avg.percentiles[p], nonZeroBn)
	}
	return avg
}

func (c *circularTxpoolStats) getGasPriceForGasAbove(gas uint64) *big.Int {
	atomic.StoreUint32(&c.activated, 1)
	avg_c := c.avg.Load()
	if avg_c == nil {
		return new(big.Int)
	}
	avg := avg_c.(txpoolStat)
	if avg.totalGas == 0 {
		return new(big.Int)
	}
	if gas > avg.totalGas {
		// special case if pool is half-empty, extrapolate linearly
		v := new(big.Int).Mul(avg.percentiles[len(avg.percentiles)-1], new(big.Int).SetUint64(avg.totalGas))
		v.Div(v, new(big.Int).SetUint64(gas+1))
		return v
	}
	p := gas * uint64(len(avg.percentiles)) / avg.totalGas
	if p >= uint64(len(avg.percentiles)) {
		p = uint64(len(avg.percentiles)) - 1
	}
	return avg.percentiles[p]
}

func (c *circularTxpoolStats) totalGas() uint64 {
	atomic.StoreUint32(&c.activated, 1)
	avgC := c.avg.Load()
	if avgC == nil {
		return 0
	}
	avg := avgC.(txpoolStat)
	return avg.totalGas
}

// calcTxpoolStat retrieves txpool transactions and calculates statistics
func (gpo *Oracle) calcTxpoolStat() txpoolStat {
	txs := gpo.backend.PendingTxs()
	s := txpoolStat{}
	if len(txs) == 0 {
		// short circuit if empty txpool
		return s
	}
	// don't index more transactions than needed for GPO purposes
	const maxTxsToIndex = 400
	const maxGasToIndex = 50000000

	minGasPrice := gpo.backend.GetRules().Economy.MinGasPrice
	// txs are sorted from large price to small
	sorted := txs
	sort.Slice(sorted, func(i, j int) bool {
		a, b := sorted[i], sorted[j]
		return a.EffectiveGasTipCmp(b, minGasPrice) < 0
	})

	if len(txs) > maxTxsToIndex {
		txs = txs[:maxTxsToIndex]
	}
	sortedDown := make(types.Transactions, len(sorted))
	for i, tx := range sorted {
		sortedDown[len(sorted)-1-i] = tx
	}

	for i, tx := range sortedDown {
		s.totalGas += tx.Gas()
		if s.totalGas > maxGasToIndex {
			sortedDown = sortedDown[:i+1]
			break
		}
	}

	gasCounter := uint64(0)
	p := uint64(0)
	for _, tx := range sortedDown {
		gasCounter += tx.Gas()
		for p < uint64(len(s.percentiles)) && gasCounter >= p*s.totalGas/uint64(len(s.percentiles)) {
			s.percentiles[p] = tx.EffectiveGasTipValue(minGasPrice)
			if s.percentiles[p].Sign() < 0 {
				s.percentiles[p] = minGasPrice
			} else {
				s.percentiles[p].Add(s.percentiles[p], minGasPrice)
			}
			p++
		}
	}

	return s
}
