package emitter

import (
	"time"

	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/inter/pos"

	"github.com/Fantom-foundation/go-opera/gossip/emitter/piecefunc"
)

const (
	spareValidatorThreshold = 0.8 * piecefunc.DecimalUnit
	validatorChallenge      = 5 * time.Second
)

func (em *Emitter) recountValidators(validators *pos.Validators) {
	// stakers with lower stake should emit less events to reduce network load
	// confirmingEmitInterval = piecefunc(totalStakeBeforeMe / totalStake) * MinEmitInterval
	myIdx := validators.GetIdx(em.config.Validator.ID)
	totalStakeBefore := pos.Weight(0)
	myStakeRatio := uint64(0)
	for i, stake := range validators.SortedWeights() {
		vid := validators.GetID(idx.Validator(i))
		// pos.Weight is uint32, so cast to uint64 to avoid an overflow
		stakeRatio := uint64(totalStakeBefore) * uint64(piecefunc.DecimalUnit) / uint64(validators.TotalWeight())
		if idx.Validator(i) == myIdx {
			myStakeRatio = stakeRatio
		}
		if stakeRatio > spareValidatorThreshold {
			em.spareValidators[vid] = true
		} else {
			delete(em.spareValidators, vid)
		}
		if !em.offlineValidators[vid] || idx.Validator(i) == myIdx {
			totalStakeBefore += stake
		}
	}
	confirmingEmitIntervalRatio := piecefunc.Get(myStakeRatio, confirmingEmitIntervalF)
	em.intervals.Confirming = time.Duration(piecefunc.Mul(uint64(em.config.EmitIntervals.Confirming), confirmingEmitIntervalRatio))
	em.intervals.Max = em.config.EmitIntervals.Max
}

func (em *Emitter) recheckChallenges() {
	if time.Since(em.prevRecheckedChallenges) < validatorChallenge/10 {
		return
	}
	em.world.EngineMu.Lock()
	defer em.world.EngineMu.Unlock()
	now := time.Now()
	if !em.idle() {
		// give challenges to all the non-spare validators
		validators := em.world.Store.GetValidators()
		for _, vid := range validators.IDs() {
			if em.spareValidators[vid] || em.offlineValidators[vid] {
				continue
			}
			if _, ok := em.challenges[vid]; !ok {
				em.challenges[vid] = now.Add(validatorChallenge)
			}
		}
	} else {
		// erase all the challenges
		em.challenges = make(map[idx.ValidatorID]time.Time)
	}
	// check challenges
	recount := false
	for vid, challengeDeadline := range em.challenges {
		if now.After(challengeDeadline) {
			em.offlineValidators[vid] = true
			recount = true
		}
	}
	if recount {
		em.recountValidators(em.world.Store.GetValidators())
	}
	em.prevRecheckedChallenges = now
}
