package gossip

import (
	"sync/atomic"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/inter/pos"

	"github.com/Fantom-foundation/go-opera/eventcheck/gaspowercheck"
	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/inter/validatorpk"
	"github.com/Fantom-foundation/go-opera/opera"
)

// GasPowerCheckReader is a helper to run gas power check
type GasPowerCheckReader struct {
	Ctx atomic.Value
}

// GetValidationContext returns current validation context for gaspowercheck
func (r *GasPowerCheckReader) GetValidationContext() *gaspowercheck.ValidationContext {
	return r.Ctx.Load().(*gaspowercheck.ValidationContext)
}

// NewGasPowerContext reads current validation context for gaspowercheck
func NewGasPowerContext(s *Store, validators *pos.Validators, epoch idx.Epoch, cfg *opera.EconomyConfig) *gaspowercheck.ValidationContext {
	// engineMu is locked here

	short := cfg.ShortGasPower
	shortTermConfig := gaspowercheck.Config{
		Idx:                inter.ShortTermGas,
		AllocPerSec:        short.InitialAllocPerSec,
		MaxAllocPeriod:     short.MaxAllocPeriod,
		StartupAllocPeriod: short.StartupAllocPeriod,
		MinStartupGas:      short.MinStartupGas,
	}

	long := cfg.LongGasPower
	longTermConfig := gaspowercheck.Config{
		Idx:                inter.LongTermGas,
		AllocPerSec:        long.InitialAllocPerSec,
		MaxAllocPeriod:     long.MaxAllocPeriod,
		StartupAllocPeriod: long.StartupAllocPeriod,
		MinStartupGas:      long.MinStartupGas,
	}

	validatorStates := make([]gaspowercheck.ValidatorState, validators.Len())
	es := s.GetEpochState()
	for i, val := range es.ValidatorStates {
		validatorStates[i].GasRefund = val.GasRefund
		if val.PrevEpochEvent != hash.ZeroEvent {
			validatorStates[i].PrevEpochEvent = s.GetEvent(val.PrevEpochEvent)
		}
	}

	return &gaspowercheck.ValidationContext{
		Epoch:           epoch,
		Validators:      validators,
		EpochStart:      es.EpochStart,
		ValidatorStates: validatorStates,
		Configs: [inter.GasPowerConfigs]gaspowercheck.Config{
			inter.ShortTermGas: shortTermConfig,
			inter.LongTermGas:  longTermConfig,
		},
	}
}

// ValidatorsPubKeys stores info to authenticate validators
type ValidatorsPubKeys struct {
	Epoch   idx.Epoch
	PubKeys map[idx.ValidatorID]validatorpk.PubKey
}

// HeavyCheckReader is a helper to run heavy power checks
type HeavyCheckReader struct {
	Addrs atomic.Value
}

// GetEpochPubKeys is safe for concurrent use
func (r *HeavyCheckReader) GetEpochPubKeys() (map[idx.ValidatorID]validatorpk.PubKey, idx.Epoch) {
	auth := r.Addrs.Load().(*ValidatorsPubKeys)

	return auth.PubKeys, auth.Epoch
}

// NewEpochPubKeys is the same as GetEpochValidators, but returns only addresses
func NewEpochPubKeys(s *Store, epoch idx.Epoch) *ValidatorsPubKeys {
	es := s.GetEpochState()
	pubkeys := make(map[idx.ValidatorID]validatorpk.PubKey, len(es.ValidatorProfiles))
	for id, profile := range es.ValidatorProfiles {
		pubkeys[id] = profile.PubKey
	}
	return &ValidatorsPubKeys{
		Epoch:   epoch,
		PubKeys: pubkeys,
	}
}
