package emitter

import (
	"time"

	"github.com/Fantom-foundation/lachesis-base/common/bigendian"
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/inter/pos"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"

	"github.com/Fantom-foundation/go-opera/eventcheck/epochcheck"
	"github.com/Fantom-foundation/go-opera/eventcheck/gaspowercheck"
	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/utils"
)

const (
	TxTimeBufferSize    = 20000
	TxTurnPeriod        = 8 * time.Second
	TxTurnPeriodLatency = 1 * time.Second
	TxTurnNonces        = 32
)

func max64(a, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}

func (em *Emitter) maxGasPowerToUse(e *inter.MutableEventPayload) uint64 {
	rules := em.world.GetRules()
	maxGasToUse := rules.Economy.Gas.MaxEventGas
	if maxGasToUse > e.GasPowerLeft().Min() {
		maxGasToUse = e.GasPowerLeft().Min()
	}
	// Smooth TPS if power isn't big
	if em.config.LimitedTpsThreshold > em.config.NoTxsThreshold {
		upperThreshold := em.config.LimitedTpsThreshold
		downThreshold := em.config.NoTxsThreshold

		estimatedAlloc := gaspowercheck.CalcValidatorGasPower(e, e.CreationTime(), e.MedianTime(), 0, em.validators, gaspowercheck.Config{
			Idx:                inter.LongTermGas,
			AllocPerSec:        rules.Economy.LongGasPower.AllocPerSec * 4 / 5,
			MaxAllocPeriod:     inter.Timestamp(time.Minute),
			MinEnsuredAlloc:    0,
			StartupAllocPeriod: 0,
			MinStartupGas:      0,
		})

		gasPowerLeft := e.GasPowerLeft().Min() + estimatedAlloc
		if gasPowerLeft < downThreshold {
			return 0
		}
		newGasPowerLeft := uint64(0)
		if gasPowerLeft > maxGasToUse {
			newGasPowerLeft = gasPowerLeft - maxGasToUse
		}

		var x1, x2 = newGasPowerLeft, gasPowerLeft
		if x1 < downThreshold {
			x1 = downThreshold
		}
		if x2 > upperThreshold {
			x2 = upperThreshold
		}
		trespassingPart := uint64(0)
		if x2 > x1 {
			trespassingPart = x2 - x1
		}
		healthyPart := uint64(0)
		if gasPowerLeft > x2 {
			healthyPart = gasPowerLeft - x2
		}

		smoothGasToUse := healthyPart + trespassingPart/2
		if maxGasToUse > smoothGasToUse {
			maxGasToUse = smoothGasToUse
		}
	}
	// pendingGas should be below MaxBlockGas
	{
		maxPendingGas := max64(rules.Blocks.MaxBlockGas*3/5, rules.Economy.Gas.MaxEventGas+rules.Economy.Gas.EventGas*uint64(em.validators.Len()))
		if maxPendingGas <= em.pendingGas {
			return 0
		}
		if maxPendingGas < em.pendingGas+maxGasToUse {
			maxGasToUse = maxPendingGas - em.pendingGas
		}
	}
	// No txs if power is low
	{
		threshold := em.config.NoTxsThreshold
		if e.GasPowerLeft().Min() <= threshold {
			return 0
		} else if e.GasPowerLeft().Min() < threshold+maxGasToUse {
			maxGasToUse = e.GasPowerLeft().Min() - threshold
		}
	}
	return maxGasToUse
}

// safe for concurrent use
func (em *Emitter) memorizeTxTimes(txs types.Transactions) {
	if em.config.Validator.ID == 0 {
		return // short circuit if not a validator
	}
	now := time.Now()
	for _, tx := range txs {
		if _, ok := em.txTime.Get(tx.Hash()); !ok { 
			em.txTime.Add(tx.Hash(), now)
		}
	}
}

func getTxRoundIndex(now, txTime time.Time, validatorsNum idx.Validator) int {
	passed := now.Sub(txTime)
	if passed < 0 {
		passed = 0
	}
	return int((passed / TxTurnPeriod) % time.Duration(validatorsNum))
}

func (em *Emitter) getTxTime(txHash common.Hash) time.Time {
	txTimeI, ok := em.txTime.Get(txHash)
	if !ok {
		now := time.Now()
		em.txTime.Add(txHash, now)
		return now
	}
	return txTimeI.(time.Time)
}

// safe for concurrent use
func (em *Emitter) isMyTxTurn(txHash common.Hash, sender common.Address, accountNonce uint64, now time.Time, validators *pos.Validators, me idx.ValidatorID, epoch idx.Epoch) bool {
	txTime := em.getTxTime(txHash)

	roundIndex := getTxRoundIndex(now, txTime, validators.Len())
	if roundIndex != getTxRoundIndex(now.Add(TxTurnPeriodLatency), txTime, validators.Len()) {
		// round is about to change, avoid originating the transaction to avoid racing with another validator
		return false
	}

	roundsHash := hash.Of(sender.Bytes(), bigendian.Uint64ToBytes(accountNonce/TxTurnNonces), epoch.Bytes())
	rounds := utils.WeightedPermutation(roundIndex+1, validators.SortedWeights(), roundsHash)
	return validators.GetID(idx.Validator(rounds[roundIndex])) == me
}

func (em *Emitter) addTxs(e *inter.MutableEventPayload, sorted *types.TransactionsByPriceAndNonce) {
	maxGasUsed := em.maxGasPowerToUse(e)
	if maxGasUsed <= e.GasPowerUsed() {
		return
	}

	// sort transactions by price and nonce
	for tx := sorted.Peek(); tx != nil; tx = sorted.Peek() {
		sender, _ := types.Sender(em.world.TxSigner, tx)
		// check transaction epoch rules
		if epochcheck.CheckTxs(types.Transactions{tx}, em.world.GetRules()) != nil {
			sorted.Pop()
			continue
		}
		// check there's enough gas power to originate the transaction
		if tx.Gas() >= e.GasPowerLeft().Min() || e.GasPowerUsed()+tx.Gas() >= maxGasUsed {
			if params.TxGas >= e.GasPowerLeft().Min() || e.GasPowerUsed()+params.TxGas >= maxGasUsed {
				// stop if cannot originate even an empty transaction
				break
			}
			sorted.Pop()
			continue
		}
		// check not conflicted with already originated txs (in any connected event)
		if em.originatedTxs.TotalOf(sender) != 0 {
			sorted.Pop()
			continue
		}
		// my turn, i.e. try to not include the same tx simultaneously by different validators
		if !em.isMyTxTurn(tx.Hash(), sender, tx.Nonce(), time.Now(), em.validators, e.Creator(), em.epoch) {
			sorted.Pop()
			continue
		}
		// check transaction is not outdated
		if !em.world.TxPool.Has(tx.Hash()) {
			sorted.Pop()
			continue
		}
		// add
		e.SetGasPowerUsed(e.GasPowerUsed() + tx.Gas())
		e.SetGasPowerLeft(e.GasPowerLeft().Sub(tx.Gas()))
		e.SetTxs(append(e.Txs(), tx))
		sorted.Shift()
	}
}
