package poset

import (
	"github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/inter/pos"
)

const (
	firstFrame = idx.Frame(1)
	firstEpoch = idx.Epoch(1)
)

type EpochState struct {
	// stored values
	// these values change only after a change of epoch
	EpochN     idx.Epoch
	PrevEpoch  GenesisState
	Validators *pos.Validators
}

func (p *Poset) loadEpoch() {
	p.EpochState = *p.store.GetEpoch()
	p.store.RecreateEpochDb(p.EpochState.EpochN)
}

// GetEpoch returns current epoch num to 3rd party.
func (p *Poset) GetEpoch() idx.Epoch {
	p.epochMu.Lock()
	defer p.epochMu.Unlock()

	return p.EpochN
}

// GetValidators returns validators of current epoch.
// Don't mutate validators.
func (p *Poset) GetValidators() *pos.Validators {
	p.epochMu.Lock()
	defer p.epochMu.Unlock()

	return p.Validators
}

// GetEpochValidators returns validators of current epoch, and the epoch.
// Don't mutate validators.
func (p *Poset) GetEpochValidators() (*pos.Validators, idx.Epoch) {
	p.epochMu.Lock()
	defer p.epochMu.Unlock()

	return p.Validators, p.EpochN
}

func (p *Poset) setEpochValidators(validators *pos.Validators, epoch idx.Epoch) {
	p.epochMu.Lock()
	defer p.epochMu.Unlock()

	p.Validators = validators
	p.EpochN = epoch
}

// GetGenesisHash returns PrevEpochHash of first epoch.
func (p *Poset) GetGenesisHash() common.Hash {
	return p.store.GetGenesisHash()
}
