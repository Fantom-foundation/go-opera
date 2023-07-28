package gasprice

import (
	"math/big"
	"sort"
	"sync/atomic"
	"time"

	"github.com/Fantom-foundation/lachesis-base/utils/piecefunc"
	"github.com/ethereum/go-ethereum/core/types"
)

const (
	percentilesPerStat = 20
	statUpdatePeriod   = 1 * time.Second
	statsBuffer        = int((15 * time.Second) / statUpdatePeriod)
	maxGasToIndex      = 40000000
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

var certaintyToGasAbove = piecefunc.NewFunc([]piecefunc.Dot{
	{
		X: 0,
		Y: 50000000,
	},
	{
		X: 0.2 * DecimalUnit,
		Y: 20000000,
	},
	{
		X: 0.5 * DecimalUnit,
		Y: 8000000,
	},
	{
		X: DecimalUnit,
		Y: 0,
	},
})

func (gpo *Oracle) reactiveGasPrice(certainty uint64) *big.Int {
	gasAbove := certaintyToGasAbove(certainty)

	return gpo.c.getGasPriceForGasAbove(gasAbove)
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

func (c *circularTxpoolStats) dec(v int) int {
	if v == 0 {
		return len(c.stats) - 1
	}
	return v - 1
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
			if s.percentiles[p] == nil {
				continue
			}
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

	// take maximum from previous 4 stats plus 5%
	rec := txpoolStat{}
	for p := range rec.percentiles {
		rec.percentiles[p] = new(big.Int)
	}
	recI1 := c.dec(c.i)
	recI2 := c.dec(recI1)
	recI3 := c.dec(recI2)
	recI4 := c.dec(recI3)
	for _, s := range []txpoolStat{c.stats[recI1], c.stats[recI2], c.stats[recI3], c.stats[recI4]} {
		for p := range s.percentiles {
			if s.percentiles[p] == nil {
				continue
			}
			if rec.percentiles[p].Cmp(s.percentiles[p]) < 0 {
				rec.percentiles[p].Set(s.percentiles[p])
			}
		}
	}
	// increase by 5%
	for p := range rec.percentiles {
		rec.percentiles[p].Mul(rec.percentiles[p], big.NewInt(21))
		rec.percentiles[p].Div(rec.percentiles[p], big.NewInt(20))
	}

	// return minimum from max(recent two stats * 1.05) and avg stats
	res := txpoolStat{}
	res.totalGas = avg.totalGas
	for _, s := range []txpoolStat{avg, rec} {
		for p := range s.percentiles {
			if res.percentiles[p] == nil || res.percentiles[p].Cmp(s.percentiles[p]) > 0 {
				res.percentiles[p] = s.percentiles[p]
			}
		}
	}

	return res
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
	if gas > maxGasToIndex {
		// extrapolate linearly
		v := new(big.Int).Mul(avg.percentiles[len(avg.percentiles)-1], new(big.Int).SetUint64(maxGasToIndex))
		v.Div(v, new(big.Int).SetUint64(gas+1))
		return v
	}
	p0 := gas * uint64(len(avg.percentiles)) / maxGasToIndex
	if p0 >= uint64(len(avg.percentiles))-1 {
		return avg.percentiles[len(avg.percentiles)-1]
	}
	// interpolate linearly
	p1 := p0 + 1
	x := gas
	x0, x1 := p0*maxGasToIndex/uint64(len(avg.percentiles)), p1*maxGasToIndex/uint64(len(avg.percentiles))
	y0, y1 := avg.percentiles[p0], avg.percentiles[p1]
	return div64I(addBigI(mul64N(y0, x1-x), mul64N(y1, x-x0)), x1-x0)
}

func mul64N(a *big.Int, b uint64) *big.Int {
	return new(big.Int).Mul(a, new(big.Int).SetUint64(b))
}

func div64I(a *big.Int, b uint64) *big.Int {
	return a.Div(a, new(big.Int).SetUint64(b))
}

func addBigI(a, b *big.Int) *big.Int {
	return a.Add(a, b)
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
	txsMap := gpo.backend.PendingTxs()
	s := txpoolStat{}
	if len(txsMap) == 0 {
		// short circuit if empty txpool
		return s
	}
	// take only one tx from each account
	txs := make(types.Transactions, 0, 1000)
	for _, aTxs := range txsMap {
		txs = append(txs, aTxs[0])
	}

	// don't index more transactions than needed for GPO purposes
	const maxTxsToIndex = 400

	minGasPrice := gpo.backend.GetRules().Economy.MinGasPrice
	// txs are sorted from large price to small
	sorted := txs
	sort.Slice(sorted, func(i, j int) bool {
		a, b := sorted[i], sorted[j]
		cmp := a.EffectiveGasTipCmp(b, minGasPrice)
		if cmp == 0 {
			return a.Gas() > b.Gas()
		}
		return cmp > 0
	})

	for i, tx := range sorted {
		s.totalGas += tx.Gas()
		if s.totalGas > maxGasToIndex || i > maxTxsToIndex {
			sorted = sorted[:i+1]
			break
		}
	}

	gasCounter := uint64(0)
	p := uint64(0)
	for _, tx := range sorted {
		for p < uint64(len(s.percentiles)) && gasCounter >= p*maxGasToIndex/uint64(len(s.percentiles)) {
			s.percentiles[p] = tx.EffectiveGasTipValue(minGasPrice)
			if s.percentiles[p].Sign() < 0 {
				s.percentiles[p] = minGasPrice
			} else {
				s.percentiles[p].Add(s.percentiles[p], minGasPrice)
			}
			p++
		}
		gasCounter += tx.Gas()
	}

	return s
}
