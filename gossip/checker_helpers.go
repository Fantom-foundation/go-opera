package gossip

import (
	"sync/atomic"

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
func NewGasPowerContext(s *Store, validators *pos.Validators, epoch idx.Epoch, cfg opera.EconomyRules) *gaspowercheck.ValidationContext {
	// engineMu is locked here

	short := cfg.ShortGasPower
	shortTermConfig := gaspowercheck.Config{
		Idx:                inter.ShortTermGas,
		AllocPerSec:        short.AllocPerSec,
		MaxAllocPeriod:     short.MaxAllocPeriod,
		MinEnsuredAlloc:    cfg.Gas.MaxEventGas,
		StartupAllocPeriod: short.StartupAllocPeriod,
		MinStartupGas:      short.MinStartupGas,
	}

	long := cfg.LongGasPower
	longTermConfig := gaspowercheck.Config{
		Idx:                inter.LongTermGas,
		AllocPerSec:        long.AllocPerSec,
		MaxAllocPeriod:     long.MaxAllocPeriod,
		MinEnsuredAlloc:    cfg.Gas.MaxEventGas,
		StartupAllocPeriod: long.StartupAllocPeriod,
		MinStartupGas:      long.MinStartupGas,
	}

	validatorStates := make([]gaspowercheck.ValidatorState, validators.Len())
	es := s.GetEpochState()
	for i, val := range es.ValidatorStates {
		validatorStates[i].GasRefund = val.GasRefund
		validatorStates[i].PrevEpochEvent = val.PrevEpochEvent
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
	Pubkeys atomic.Value
	Store   *Store
}

// GetEpochPubKeys is safe for concurrent use
func (r *HeavyCheckReader) GetEpochPubKeys() (map[idx.ValidatorID]validatorpk.PubKey, idx.Epoch) {
	auth := r.Pubkeys.Load().(*ValidatorsPubKeys)

	return auth.PubKeys, auth.Epoch
}

// GetEpochPubKeysOf is safe for concurrent use
func (r *HeavyCheckReader) GetEpochPubKeysOf(epoch idx.Epoch) map[idx.ValidatorID]validatorpk.PubKey {
	auth := readEpochPubKeys(r.Store, epoch)
	if auth == nil {
		return nil
	}
	return auth.PubKeys
}

// GetEpochBlockStart is safe for concurrent use
func (r *HeavyCheckReader) GetEpochBlockStart(epoch idx.Epoch) idx.Block {
	bs, _ := r.Store.GetHistoryBlockEpochState(epoch)
	if bs == nil {
		return 0
	}
	return bs.LastBlock.Idx
}

// readEpochPubKeys reads epoch pubkeys
func readEpochPubKeys(s *Store, epoch idx.Epoch) *ValidatorsPubKeys {
	es := s.GetHistoryEpochState(epoch)
	if es == nil {
		return nil
	}
	var pubkeys = make(map[idx.ValidatorID]validatorpk.PubKey, len(es.ValidatorProfiles))
	for id, profile := range es.ValidatorProfiles {
		pubkeys[id] = profile.PubKey
	}
	return &ValidatorsPubKeys{
		Epoch:   epoch,
		PubKeys: pubkeys,
	}
}
