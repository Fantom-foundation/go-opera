package emitter

import (
	"time"

	"github.com/Fantom-foundation/lachesis-base/common/bigendian"
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/inter/pos"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/opera/params"
	"github.com/Fantom-foundation/go-opera/utils"
)

const (
	TxTimeBufferSize  = 20000
	TxTurnPeriod      = 8 * time.Second
	TxTurnPeriodSlack = 1 * time.Second
	TxTurnNonces      = 24
)

func (em *Emitter) maxGasPowerToUse(e *inter.MutableEventPayload) uint64 {
	// No txs if power is low
	{
		threshold := em.config.NoTxsThreshold
		if e.GasPowerLeft().Min() <= threshold {
			return 0
		}
		if e.GasPowerLeft().Min() < threshold+params.MaxGasPowerUsed {
			return e.GasPowerLeft().Min() - threshold
		}
	}
	// Smooth TPS if power isn't big
	{
		threshold := em.config.LimitedTpsThreshold
		if e.GasPowerLeft().Min() <= threshold {
			// it's emitter, so no need in determinism => fine to use float
			passedTime := float64(e.CreationTime().Time().Sub(em.prevEmittedAtTime)) / (float64(time.Second))
			maxGasUsed := uint64(passedTime * em.gasRate.Rate1() * em.config.MaxGasRateGrowthFactor)
			if maxGasUsed > params.MaxGasPowerUsed {
				maxGasUsed = params.MaxGasPowerUsed
			}
			return maxGasUsed
		}
	}
	return params.MaxGasPowerUsed
}

// safe for concurrent use
func (em *Emitter) memorizeTxTimes(txs types.Transactions) {
	if em.config.Validator.ID == 0 {
		return // short circuit if not a validator
	}
	now := time.Now()
	for _, tx := range txs {
		_, ok := em.txTime.Get(tx.Hash())
		if !ok {
			em.txTime.Add(tx.Hash(), now)
		}
	}
}

// safe for concurrent use
func (em *Emitter) isMyTxTurn(txHash common.Hash, sender common.Address, accountNonce uint64, now time.Time, validatorsArr []idx.ValidatorID, validatorsArrStakes []pos.Weight, me idx.ValidatorID, epoch idx.Epoch) bool {

	var txTime time.Time
	txTimeI, ok := em.txTime.Get(txHash)
	if !ok {
		txTime = now
		em.txTime.Add(txHash, txTime)
	} else {
		txTime = txTimeI.(time.Time)
	}

	getRoundIndex := func(t time.Time) int {
		return int((t.Sub(txTime) / TxTurnPeriod) % time.Duration(len(validatorsArr)))
	}
	roundIndex := getRoundIndex(now)
	if roundIndex != getRoundIndex(now.Add(TxTurnPeriodSlack)) {
		return false
	}

	turnHash := hash.Of(sender.Bytes(), bigendian.Uint64ToBytes(accountNonce/TxTurnNonces), epoch.Bytes())

	turns := utils.WeightedPermutation(roundIndex+1, validatorsArrStakes, turnHash)

	return validatorsArr[turns[roundIndex]] == me
}

func (em *Emitter) addTxs(e *inter.MutableEventPayload, poolTxs map[common.Address]types.Transactions) {
	if len(poolTxs) == 0 {
		return
	}

	maxGasUsed := em.maxGasPowerToUse(e)

	validators, epoch := em.world.Store.GetEpochValidators()
	validatorsArr := validators.SortedIDs() // validators must be sorted deterministically
	validatorsArrStakes := make([]pos.Weight, len(validatorsArr))
	for i, addr := range validatorsArr {
		validatorsArrStakes[i] = validators.Get(addr)
	}

	sorted := types.NewTransactionsByPriceAndNonce(em.world.TxSigner, poolTxs)

	senderTxs := make(map[common.Address]int)
	for tx := sorted.Peek(); tx != nil; tx = sorted.Peek() {
		sender, _ := types.Sender(em.world.TxSigner, tx)
		if senderTxs[sender] >= em.config.MaxTxsPerAddress {
			sorted.Pop()
			continue
		}
		if tx.Gas() >= e.GasPowerLeft().Min() || e.GasPowerUsed()+tx.Gas() >= maxGasUsed {
			sorted.Pop()
			continue
		}
		// check not conflicted with already included txs (in any connected event)
		if em.originatedTxs.TotalOf(sender) != 0 {
			sorted.Pop()
			continue
		}
		// my turn, i.e. try to not include the same tx simultaneously by different validators
		if !em.isMyTxTurn(tx.Hash(), sender, tx.Nonce(), time.Now(), validatorsArr, validatorsArrStakes, e.Creator(), epoch) {
			sorted.Pop()
			continue
		}
		// check transaction is not outdated
		if !em.world.Txpool.Has(tx.Hash()) {
			sorted.Pop()
			continue
		}
		// add
		e.SetGasPowerUsed(e.GasPowerUsed() + tx.Gas())
		e.SetGasPowerLeft(e.GasPowerLeft().Sub(tx.Gas()))
		e.SetTxs(append(e.Txs(), tx))
		senderTxs[sender]++
		sorted.Shift()
	}
}
