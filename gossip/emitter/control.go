package emitter

import (
	"time"

	"github.com/Fantom-foundation/lachesis-base/emitter/ancestor"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/inter/pos"

	"github.com/Fantom-foundation/go-opera/gossip/emitter/piecefunc"
	"github.com/Fantom-foundation/go-opera/inter"
)

func scalarUpdMetric(diff idx.Event, weight pos.Weight, totalWeight pos.Weight) ancestor.Metric {
	return ancestor.Metric(piecefunc.Get(uint64(diff)*piecefunc.DecimalUnit, scalarUpdMetricF)) * ancestor.Metric(weight) / ancestor.Metric(totalWeight)
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
	metric := ancestor.Metric(piecefunc.Get(uint64(orig), eventMetricF))
	// kick start metric in a beginning of epoch, when there's nothing to observe yet
	if seq <= 2 && metric < 0.9*piecefunc.DecimalUnit {
		metric += 0.1 * piecefunc.DecimalUnit
	}
	if seq <= 1 && metric <= 0.8*piecefunc.DecimalUnit {
		metric += 0.2 * piecefunc.DecimalUnit
	}
	return metric
}

func (em *Emitter) isAllowedToEmit(e inter.EventPayloadI, metric ancestor.Metric, selfParent *inter.Event) bool {
	passedTime := e.CreationTime().Time().Sub(em.prevEmittedAtTime)
	adjustedPassedTime := time.Duration(ancestor.Metric(passedTime/piecefunc.DecimalUnit) * metric)
	passedBlocks := em.world.Store.GetLatestBlockIndex() - em.prevEmittedAtBlock
	// Forbid emitting if not enough power and power is decreasing
	{
		threshold := em.config.EmergencyThreshold
		if e.GasPowerLeft().Min() <= threshold {
			if selfParent != nil && e.GasPowerLeft().Min() < selfParent.GasPowerLeft().Min() {
				validators := em.world.Store.GetValidators()
				em.Periodic.Warn(10*time.Second, "Not enough power to emit event, waiting",
					"power", e.GasPowerLeft().String(),
					"selfParentPower", selfParent.GasPowerLeft().String(),
					"stake%", 100*float64(validators.Get(e.Creator()))/float64(validators.TotalWeight()))
				return false
			}
		}
	}
	// Enforce emitting if passed too many time/blocks since previous event
	{
		maxBlocks := em.net.Economy.BlockMissedSlack/2 + 1
		if em.net.Economy.BlockMissedSlack > maxBlocks && maxBlocks < em.net.Economy.BlockMissedSlack-5 {
			maxBlocks = em.net.Economy.BlockMissedSlack - 5
		}
		if passedTime >= em.intervals.Max ||
			passedBlocks >= maxBlocks*4/5 && metric >= piecefunc.DecimalUnit/2 ||
			passedBlocks >= maxBlocks {
			return true
		}
	}
	// Slow down emitting if power is low
	{
		threshold := em.config.NoTxsThreshold
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
			len(e.Txs()) == 0 {
			return false
		}
	}
	// Emitting is controlled by the efficiency metric
	{
		if passedTime < em.intervals.Min {
			return false
		}
		if !em.idle() && adjustedPassedTime < em.intervals.Min {
			return false
		}
		if adjustedPassedTime < em.intervals.Confirming &&
			!em.idle() &&
			len(e.Txs()) == 0 {
			return false
		}
	}

	return true
}
