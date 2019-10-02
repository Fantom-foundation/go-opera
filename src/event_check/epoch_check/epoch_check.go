package epoch_check

import (
	"errors"

	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/inter/pos"
	"github.com/Fantom-foundation/go-lachesis/src/lachesis"
)

var (
	ErrNotRelevant = errors.New("event is too old or too new")
	ErrAuth      = errors.New("event creator isn't team member")
)

type DagReader interface {
	GetEpochMembers() (pos.Members, idx.Epoch)
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
	members, epoch := v.reader.GetEpochMembers()
	if e.Epoch != epoch {
		return ErrNotRelevant
	}
	if _, ok := members[e.Creator]; !ok {
		return ErrAuth
	}
	return nil
}
