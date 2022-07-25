package emitter

import (
	"time"

	"github.com/Fantom-foundation/lachesis-base/emitter/ancestor"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/inter/pos"
	"github.com/Fantom-foundation/lachesis-base/utils/piecefunc"

	"github.com/Fantom-foundation/go-opera/inter"
)

func scalarUpdMetric(diff idx.Event, weight pos.Weight, totalWeight pos.Weight) ancestor.Metric {
	return ancestor.Metric(scalarUpdMetricF(uint64(diff)*piecefunc.DecimalUnit)) * ancestor.Metric(weight) / ancestor.Metric(totalWeight)
}

func updMetric(median, cur, upd idx.Event, validatorIdx idx.Validator, validators *pos.Validators) ancestor.Metric {
	if upd <= median || upd <= cur {
		return 0
	}
	weight := validators.GetWeightByIdx(validatorIdx)
	if median < cur {
		return scalarUpdMetric(upd-median, weight, validators.TotalWeight()) - scalarUpdMetric(cur-median, weight, validators.TotalWeight())
	}
	return scalarUpdMetric(upd-median, weight, validators.TotalWeight())
}

func eventMetric(orig ancestor.Metric, seq idx.Event) ancestor.Metric {
	metric := ancestor.Metric(eventMetricF(uint64(orig)))
	// kick start metric in a beginning of epoch, when there's nothing to observe yet
	if seq <= 2 && metric < 0.9*piecefunc.DecimalUnit {
		metric += 0.1 * piecefunc.DecimalUnit
	}
	if seq <= 1 && metric <= 0.8*piecefunc.DecimalUnit {
		metric += 0.2 * piecefunc.DecimalUnit
	}
	return metric
}

func (em *Emitter) isAllowedToEmit(e inter.EventI, eTxs bool, metric ancestor.Metric, selfParent *inter.Event) bool {
	passedTime := e.CreationTime().Time().Sub(em.prevEmittedAtTime)
	if passedTime < 0 {
		passedTime = 0
	}
	passedTimeIdle := e.CreationTime().Time().Sub(em.prevIdleTime)
	if passedTimeIdle < 0 {
		passedTimeIdle = 0
	}
	if em.stakeRatio[e.Creator()] < 0.35*piecefunc.DecimalUnit {
		// top validators emit event right after transaction is originated
		passedTimeIdle = passedTime
	} else if em.stakeRatio[e.Creator()] < 0.7*piecefunc.DecimalUnit {
		// top validators emit event right after transaction is originated
		passedTimeIdle = (passedTimeIdle + passedTime) / 2
	}
	if passedTimeIdle > passedTime {
		passedTimeIdle = passedTime
	}
	// metric is a decimal (0.0, 1.0], being an estimation of how much the event will advance the consensus
	adjustedPassedTime := time.Duration(ancestor.Metric(passedTime/piecefunc.DecimalUnit) * metric)
	adjustedPassedIdleTime := time.Duration(ancestor.Metric(passedTimeIdle/piecefunc.DecimalUnit) * metric)
	passedBlocks := em.world.GetLatestBlockIndex() - em.prevEmittedAtBlock
	// Forbid emitting if not enough power and power is decreasing
	{
		threshold := em.config.EmergencyThreshold
		if e.GasPowerLeft().Min() <= threshold {
			if selfParent != nil && e.GasPowerLeft().Min() < selfParent.GasPowerLeft().Min() {
				em.Periodic.Warn(10*time.Second, "Not enough power to emit event, waiting",
					"power", e.GasPowerLeft().String(),
					"selfParentPower", selfParent.GasPowerLeft().String(),
					"stake%", 100*float64(em.validators.Get(e.Creator()))/float64(em.validators.TotalWeight()))
				return false
			}
		}
	}
	// Enforce emitting if passed too many time/blocks since previous event
	{
		rules := em.world.GetRules()
		maxBlocks := rules.Economy.BlockMissedSlack/2 + 1
		if rules.Economy.BlockMissedSlack > maxBlocks && maxBlocks < rules.Economy.BlockMissedSlack-5 {
			maxBlocks = rules.Economy.BlockMissedSlack - 5
		}
		if passedTime >= em.intervals.Max ||
			passedBlocks >= maxBlocks*4/5 && metric >= piecefunc.DecimalUnit/2 ||
			passedBlocks >= maxBlocks {
			return true
		}
	}
	// Slow down emitting if power is low
	{
		threshold := (em.config.NoTxsThreshold + em.config.EmergencyThreshold) / 2
		if e.GasPowerLeft().Min() <= threshold {
			// it's emitter, so no need in determinism => fine to use float
			minT := float64(em.intervals.Min)
			maxT := float64(em.intervals.Max)
			factor := float64(e.GasPowerLeft().Min()) / float64(threshold)
			adjustedEmitInterval := time.Duration(maxT - (maxT-minT)*factor)
			if passedTime < adjustedEmitInterval {
				return false
			}
		}
	}
	// Slow down emitting if no txs to confirm/originate
	{
		if passedTime < em.intervals.Max &&
			em.idle() &&
			!eTxs {
			return false
		}
	}
	// Emitting is controlled by the efficiency metric
	{
		if passedTime < em.intervals.Min {
			return false
		}
		if adjustedPassedTime < em.intervals.Min &&
			!em.idle() {
			return false
		}
		if adjustedPassedIdleTime < em.intervals.Confirming &&
			!em.idle() &&
			!eTxs {
			return false
		}
	}

	return true
}

func (em *Emitter) recheckIdleTime() {
	em.world.Lock()
	defer em.world.Unlock()
	if em.idle() {
		em.prevIdleTime = time.Now()
	}
}
