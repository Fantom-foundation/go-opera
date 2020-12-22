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
	TxTimeBufferSize    = 20000
	TxTurnPeriod        = 8 * time.Second
	TxTurnPeriodLatency = 1 * time.Second
	TxTurnNonces        = 32
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

func getTxRoundIndex(now, txTime time.Time, validators idx.Validator) int {
	return int((now.Sub(txTime) / TxTurnPeriod) % time.Duration(validators))
}

func (em *Emitter) getTxTime(txHash common.Hash) time.Time {
	txTimeI, ok := em.txTime.Get(txHash)
	if !ok {
		now := time.Now()
		em.txTime.Add(txHash, now)
		return now
	} else {
		return txTimeI.(time.Time)
	}
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

func (em *Emitter) addTxs(e *inter.MutableEventPayload, poolTxs map[common.Address]types.Transactions) {
	if len(poolTxs) == 0 {
		return
	}

	maxGasUsed := em.maxGasPowerToUse(e)

	validators, epoch := em.world.Store.GetEpochValidators()

	// sort transactions by price and nonce
	sorted := types.NewTransactionsByPriceAndNonce(em.world.TxSigner, poolTxs)

	senderTxs := make(map[common.Address]int)
	for tx := sorted.Peek(); tx != nil; tx = sorted.Peek() {
		sender, _ := types.Sender(em.world.TxSigner, tx)
		if senderTxs[sender] >= em.config.MaxTxsPerAddress {
			sorted.Pop()
			continue
		}
		// check there's enough gas power to originate the transaction
		if tx.Gas() >= e.GasPowerLeft().Min() || e.GasPowerUsed()+tx.Gas() >= maxGasUsed {
			sorted.Pop()
			continue
		}
		// check not conflicted with already originated txs (in any connected event)
		if em.originatedTxs.TotalOf(sender) != 0 {
			sorted.Pop()
			continue
		}
		// my turn, i.e. try to not include the same tx simultaneously by different validators
		if !em.isMyTxTurn(tx.Hash(), sender, tx.Nonce(), time.Now(), validators, e.Creator(), epoch) {
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
