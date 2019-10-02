package epoch_check

import (
	"errors"

	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/inter/pos"
	"github.com/Fantom-foundation/go-lachesis/lachesis"
)

var (
	ErrNotRelevant = errors.New("event is too old or too new")
	ErrAuth        = errors.New("event creator isn't validator")
)

type DagReader interface {
	GetEpochValidators() (pos.Validators, idx.Epoch)
}

// Check which require only current epoch info
type Validator struct {
	config *lachesis.DagConfig
	reader DagReader
}

func New(config *lachesis.DagConfig, reader DagReader) *Validator {
	return &Validator{
		config: config,
		reader: reader,
	}
}

func (v *Validator) Validate(e *inter.Event) error {
	// check epoch first, because validators group is known only for the current epoch
	validators, epoch := v.reader.GetEpochValidators()
	if e.Epoch != epoch {
		return ErrNotRelevant
	}
	if _, ok := validators[e.Creator]; !ok {
		return ErrAuth
	}
	return nil
}
