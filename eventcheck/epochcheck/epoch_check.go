package epochcheck

import (
	"errors"

	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/inter/pos"
	"github.com/Fantom-foundation/go-lachesis/lachesis"
)

var (
	// ErrNotRelevant indicates the event's epoch isn't equal to current epoch.
	ErrNotRelevant = errors.New("event is too old or too new")
	// ErrAuth indicates that event's creator isn't authorized to create events in current epoch.
	ErrAuth = errors.New("event creator isn't validator")
)

// DagReader is accessed by the validator to get the current state.
type DagReader interface {
	GetEpochValidators() (*pos.Validators, idx.Epoch)
}

// Checker which require only current epoch info
type Checker struct {
	config *lachesis.DagConfig
	reader DagReader
}

func New(config *lachesis.DagConfig, reader DagReader) *Checker {
	return &Checker{
		config: config,
		reader: reader,
	}
}

// Validate event
func (v *Checker) Validate(e *inter.Event) error {
	// check epoch first, because validators group is known only for the current epoch
	validators, epoch := v.reader.GetEpochValidators()
	if e.Epoch != epoch {
		return ErrNotRelevant
	}
	if !validators.Exists(e.Creator) {
		return ErrAuth
	}
	return nil
}
