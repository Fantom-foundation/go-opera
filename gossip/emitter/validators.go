package emitter

import (
	"time"

	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/inter/pos"
	"github.com/Fantom-foundation/lachesis-base/utils/piecefunc"
)

const (
	validatorChallenge = 4 * time.Second
	networkStartPeriod = 3 * time.Hour
)

func (em *Emitter) recountValidators(validators *pos.Validators) {
	// stakers with lower stake should emit less events to reduce network load
	// confirmingEmitInterval = piecefunc(totalStakeBeforeMe / totalStake) * MinEmitInterval
	totalStakeBefore := pos.Weight(0)
	for i, stake := range validators.SortedWeights() {
		vid := validators.GetID(idx.Validator(i))
		// pos.Weight is uint32, so cast to uint64 to avoid an overflow
		stakeRatio := uint64(totalStakeBefore) * uint64(piecefunc.DecimalUnit) / uint64(validators.TotalWeight())
		if !em.offlineValidators[vid] {
			totalStakeBefore += stake
		}
		confirmingEmitIntervalRatio := confirmingEmitIntervalF(stakeRatio)
		em.stakeRatio[vid] = stakeRatio
		em.expectedEmitIntervals[vid] = time.Duration(piecefunc.Mul(uint64(em.config.EmitIntervals.Confirming), confirmingEmitIntervalRatio))
	}
	em.intervals.Confirming = em.expectedEmitIntervals[em.config.Validator.ID]
	em.intervals.Max = em.config.EmitIntervals.Max
	// if network just has started, then relax the doublesign protection
	if time.Since(em.world.GetGenesisTime().Time()) < networkStartPeriod {
		em.intervals.Max /= 6
		em.intervals.DoublesignProtection /= 6
	}
}

func (em *Emitter) recheckChallenges() {
	if time.Since(em.prevRecheckedChallenges) < validatorChallenge/10 {
		return
	}
	em.world.Lock()
	defer em.world.Unlock()
	now := time.Now()
	if !em.idle() {
		// give challenges to all the non-spare validators if network isn't idle
		for _, vid := range em.validators.IDs() {
			if em.offlineValidators[vid] {
				continue
			}
			if _, ok := em.challenges[vid]; !ok {
				em.challenges[vid] = now.Add(validatorChallenge + em.expectedEmitIntervals[vid]*4)
			}
		}
	} else {
		// erase all the challenges if network is idle
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
		em.recountValidators(em.validators)
	}
	em.prevRecheckedChallenges = now
}
