package gossip

import (
	"sync/atomic"

	"github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/go-lachesis/app"
	"github.com/Fantom-foundation/go-lachesis/eventcheck/gaspowercheck"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/inter/pos"
	"github.com/Fantom-foundation/go-lachesis/lachesis"
)

// GasPowerCheckReader is a helper to run gas power check
type GasPowerCheckReader struct {
	Ctx atomic.Value
}

// GetValidationContext returns current validation context for gaspowercheck
func (r *GasPowerCheckReader) GetValidationContext() *gaspowercheck.ValidationContext {
	return r.Ctx.Load().(*gaspowercheck.ValidationContext)
}

func gasPowerBounds(initialAlloc, minAlloc, maxAlloc, customAlloc uint64) uint64 {
	allocPerSec := initialAlloc
	if customAlloc != 0 {
		allocPerSec = customAlloc
	}
	if allocPerSec > maxAlloc {
		allocPerSec = maxAlloc
	}
	if allocPerSec < minAlloc {
		allocPerSec = minAlloc
	}

	return allocPerSec
}

// ReadGasPowerContext reads current validation context for gaspowercheck
func ReadGasPowerContext(s *Store, a *app.Store, validators *pos.Validators, epoch idx.Epoch, cfg *lachesis.EconomyConfig) *gaspowercheck.ValidationContext {
	// engineMu is locked here
	sfcConstants := a.GetSfcConstants(epoch - 1)

	short := cfg.ShortGasPower
	shortAllocPerSec := gasPowerBounds(short.InitialAllocPerSec, short.MinAllocPerSec, short.MaxAllocPerSec, sfcConstants.ShortGasPowerAllocPerSec)

	shortTermConfig := gaspowercheck.Config{
		Idx:                idx.ShortTermGas,
		AllocPerSec:        shortAllocPerSec,
		MaxAllocPeriod:     short.MaxAllocPeriod,
		StartupAllocPeriod: short.StartupAllocPeriod,
		MinStartupGas:      short.MinStartupGas,
	}

	long := cfg.LongGasPower
	longAllocPerSec := gasPowerBounds(long.InitialAllocPerSec, long.MinAllocPerSec, long.MaxAllocPerSec, sfcConstants.LongGasPowerAllocPerSec)

	longTermConfig := gaspowercheck.Config{
		Idx:                idx.LongTermGas,
		AllocPerSec:        longAllocPerSec,
		MaxAllocPeriod:     long.MaxAllocPeriod,
		StartupAllocPeriod: long.StartupAllocPeriod,
		MinStartupGas:      long.MinStartupGas,
	}

	return &gaspowercheck.ValidationContext{
		Epoch:                epoch,
		Validators:           validators,
		PrevEpochLastHeaders: s.GetLastHeaders(epoch - 1),
		PrevEpochEndTime:     s.GetEpochStats(epoch - 1).End,
		PrevEpochRefunds:     a.GetGasPowerRefunds(epoch - 1),
		Configs: [2]gaspowercheck.Config{
			idx.ShortTermGas: shortTermConfig,
			idx.LongTermGas:  longTermConfig,
		},
	}
}

// ValidatorsPubKeys stores info to authenticate validators
type ValidatorsPubKeys struct {
	Epoch     idx.Epoch
	Addresses map[idx.StakerID]common.Address
}

// HeavyCheckReader is a helper to run heavy power checks
type HeavyCheckReader struct {
	Addrs atomic.Value
}

// GetEpochPubKeys is safe for concurrent use
func (r *HeavyCheckReader) GetEpochPubKeys() (map[idx.StakerID]common.Address, idx.Epoch) {
	auth := r.Addrs.Load().(*ValidatorsPubKeys)

	return auth.Addresses, auth.Epoch
}

// ReadEpochPubKeys is the same as GetEpochValidators, but returns only addresses
func ReadEpochPubKeys(a *app.Store, epoch idx.Epoch) *ValidatorsPubKeys {
	addrs := make(map[idx.StakerID]common.Address)
	for _, it := range a.GetEpochValidators(epoch) {
		addrs[it.StakerID] = it.Staker.Address
	}
	return &ValidatorsPubKeys{
		Epoch:     epoch,
		Addresses: addrs,
	}
}
