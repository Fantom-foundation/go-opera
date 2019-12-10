package gossip

import (
	"sync/atomic"

	"github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
)

func (s *Service) gasPowerCheck(e *inter.Event) error {
	// s.engineMu is locked here

	var selfParent *inter.EventHeaderData
	if e.SelfParent() != nil {
		selfParent = s.store.GetEventHeader(e.Epoch, *e.SelfParent())
	}
	return s.checkers.Gaspowercheck.Validate(e, selfParent)
}

// GasPowerCheckReader is a helper to run gas power check
type GasPowerCheckReader struct {
	Consensus
	store *Store
}

// GetPrevEpochLastHeaders isn't safe for concurrent use
func (r *GasPowerCheckReader) GetPrevEpochLastHeaders() (inter.HeadersByCreator, idx.Epoch) {
	// engineMu is locked here
	epoch := r.GetEpoch() - 1
	return r.store.GetLastHeaders(epoch), epoch
}

// GetPrevEpochEndTime isn't safe for concurrent use
func (r *GasPowerCheckReader) GetPrevEpochEndTime() (inter.Timestamp, idx.Epoch) {
	// engineMu is locked here
	epoch := r.GetEpoch() - 1
	return r.store.GetEpochStats(epoch).End, epoch
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
func (s *Store) ReadEpochPubKeys(epoch idx.Epoch) *ValidatorsPubKeys {
	addrs := make(map[idx.StakerID]common.Address)
	for _, it := range s.GetEpochValidators(epoch) {
		addrs[it.StakerID] = it.Staker.Address
	}
	return &ValidatorsPubKeys{
		Epoch:     epoch,
		Addresses: addrs,
	}
}
