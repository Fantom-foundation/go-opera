package emitter

import (
	"time"
	"fmt"
    	"math/rand"
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
	"github.com/Fantom-foundation/go-opera/utils/txtime"
)

const (
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
	//fmt.Println("maxGasToUse in maxGasPowertoUse, GasPowerLeft ", maxGasToUse, e.GasPowerLeft().Min())
	if maxGasToUse > e.GasPowerLeft().Min() {
		maxGasToUse = e.GasPowerLeft().Min()
	}
	//fmt.Println("maxGasToUse in maxGasPowertoUsem after control: ", maxGasToUse)
	// Smooth TPS if power isn't big
	if em.config.LimitedTpsThreshold > em.config.NoTxsThreshold {
		upperThreshold := em.config.LimitedTpsThreshold
		downThreshold := em.config.NoTxsThreshold

		estimatedAlloc := gaspowercheck.CalcValidatorGasPower(e, e.CreationTime(), e.MedianTime(), 0, em.validators, gaspowercheck.Config{
			Idx:                inter.LongTermGas,
			AllocPerSec:        rules.Economy.LongGasPower.AllocPerSec * 5,
			MaxAllocPeriod:     inter.Timestamp(time.Minute) * 2,
			MinEnsuredAlloc:    0,
			StartupAllocPeriod: 0,
			MinStartupGas:      0,
		})

		//fmt.Println("EstimatedAlloc: ", estimatedAlloc)
		gasPowerLeft := e.GasPowerLeft().Min() + estimatedAlloc
		//if gasPowerLeft < downThreshold {
		//	return 0
		//}
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
		//override 
		//maxGasToUse = estimatedAlloc
	}
	// pendingGas should be below MaxBlockGas
	//{
	//	maxPendingGas := max64(max64(rules.Blocks.MaxBlockGas/3, rules.Economy.Gas.MaxEventGas), 15000000)
	//	if maxPendingGas <= em.pendingGas {
	//		return 0
	//	}
		// this is likely a bug, there is no pendingGas
		//if maxPendingGas < em.pendingGas+maxGasToUse {
		//	maxGasToUse = maxPendingGas - em.pendingGas
		//}
	//}
	// No txs if power is low
	//{
	//	threshold := em.config.NoTxsThreshold
	//	if e.GasPowerLeft().Min() <= threshold {
	//		return 0
	//	} else if e.GasPowerLeft().Min() < threshold+maxGasToUse {
	//		maxGasToUse = e.GasPowerLeft().Min() - threshold
	//		fmt.Println("Show me maxGasToUse = e.GasPowerLeft().Min() - threshold ", maxGasToUse , e.GasPowerLeft().Min(), threshold)
	//	}
	//}
	maxGasToUse = rules.Economy.Gas.MaxEventGas 
	return maxGasToUse
}

// randomWithWeight returns a random number between 1 and 50,
// with the first 21 numbers having a 70% chance of being chosen.
func randomWithWeight() int {
    rand.Seed(time.Now().UnixNano()) // Initialize the random number generator.

    if rand.Float64() < 0.7 {
        // 70% chance to choose a number between 1 and 21.
        return rand.Intn(21) + 1
    } else {
        // 30% chance to choose a number between 22 and 50.
        return rand.Intn(29) + 22
    }
}

 
func getTxRoundIndex(now, txTime time.Time, validatorsNum idx.Validator) int {
       passed := now.Sub(txTime)
       if passed < 0 {
               passed = 0
       }
       return int((passed / TxTurnPeriod) % time.Duration(validatorsNum))
       }

// safe for concurrent use
func (em *Emitter) isMyTxTurn(txHash common.Hash, sender common.Address, accountNonce uint64, now time.Time, validators *pos.Validators, me idx.ValidatorID, epoch idx.Epoch) bool {
	txTime := txtime.Of(txHash)

	//roundIndex := getTxRoundIndex(now, txTime, validators.Len())
	roundIndex := getTxRoundIndex(now, txTime, 50)
	if roundIndex != getTxRoundIndex(now.Add(TxTurnPeriodLatency), txTime, validators.Len()) {
		// round is about to change, avoid originating the transaction to avoid racing with another validator
		return false
		//return true
	}

	roundsHash := hash.Of(sender.Bytes(), bigendian.Uint64ToBytes(accountNonce/TxTurnNonces), epoch.Bytes())
	rounds := utils.WeightedPermutation(roundIndex+1, validators.SortedWeights(), roundsHash)
	fmt.Println ("Validator turn id: ", rounds, validators.GetID(idx.Validator(rounds[roundIndex])))
	chosenVal := randomWithWeight()
	fmt.Println ("Validator new turn id: ", chosenVal, chosenVal ==  int(me))
	return true
	//return validators.GetID(idx.Validator(rounds[roundIndex])) == me
}

func (em *Emitter) addTxs(e *inter.MutableEventPayload, sorted *types.TransactionsByPriceAndNonce) {
	maxGasUsed := em.maxGasPowerToUse(e)
	//fmt.Println ("Old maxGasused: ",  maxGasUsed)
	oldMaxGasUsed := maxGasUsed
	if maxGasUsed <= e.GasPowerUsed() {
		fmt.Println ("Too much gas already trying to process: ",  maxGasUsed, e.GasPowerUsed())
		return
	}

	// sort transactions by price and nonce
	rules := em.world.GetRules()
	for tx := sorted.Peek(); tx != nil; tx = sorted.Peek() {
		sender, _ := types.Sender(em.world.TxSigner, tx)
		//fmt.Println("Considering TX from sender: ", sender, tx.Hash(), tx.Nonce(), time.Now(), e.Creator())
		// check transaction epoch rules
		if epochcheck.CheckTxs(types.Transactions{tx}, rules) != nil {
		        fmt.Println("Failing rules TX from sender: ", sender, tx.Hash(), tx.Nonce(), time.Now(), e.Creator())
			sorted.Pop()
			continue
		}
		// check there's enough gas power to originate the transaction
		//if tx.Gas() >= e.GasPowerLeft().Min() || e.GasPowerUsed()+tx.Gas() >= maxGasUsed {
		//	if params.TxGas >= e.GasPowerLeft().Min() || e.GasPowerUsed()+params.TxGas >= maxGasUsed {
		if e.GasPowerUsed()+tx.Gas() >= maxGasUsed {
			if e.GasPowerUsed()+params.TxGas >= maxGasUsed {

				// stop if cannot originate even an empty transaction
				break
			}
		        fmt.Println("Gas Issues TX from sender: ", sender, tx.Hash(), tx.Nonce(), time.Now(),  e.Creator() , tx.Gas() , e.GasPowerLeft().Min(),  e.GasPowerUsed()+tx.Gas(), maxGasUsed)
			sorted.Pop()
			continue
		}
		// check not conflicted with already originated txs (in any connected event)
		if em.originatedTxs.TotalOf(sender) != 0 {
			sorted.Pop()
		        fmt.Println("Already Originated TX from sender: ", sender, tx.Hash(), tx.Nonce(), time.Now(),e.Creator())
			continue
		}
		// my turn, i.e. try to not include the same tx simultaneously by different validators
		if !em.isMyTxTurn(tx.Hash(), sender, tx.Nonce(), time.Now(), em.validators, e.Creator(), em.epoch) {
		        fmt.Println("Not my turn TX from sender: ", sender, tx.Hash(), tx.Nonce(), time.Now(), e.Creator())
			sorted.Pop()
			continue
		}
		// check transaction is not outdated
		if !em.world.TxPool.Has(tx.Hash()) {
		        fmt.Println("Outdated TX from sender: ", sender, tx.Hash(), tx.Nonce(), time.Now(), e.Creator())
			sorted.Pop()
			continue
		}
		// add
		fmt.Println("My Turn to Execute  e.GasPowerUsed()+tx.Gas() >=  maxGasUsed: ", e.GasPowerUsed()+tx.Gas(), maxGasUsed, oldMaxGasUsed) 
		e.SetGasPowerUsed(e.GasPowerUsed() + tx.Gas())
		e.SetGasPowerLeft(e.GasPowerLeft().Sub(tx.Gas()))
		e.SetTxs(append(e.Txs(), tx))
		sorted.Shift()
	}
}
