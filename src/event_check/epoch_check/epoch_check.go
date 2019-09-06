package epoch_check

import (
	"errors"

	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/inter/pos"
	"github.com/Fantom-foundation/go-lachesis/src/lachesis"
)

var (
	ErrNotRecent = errors.New("event is too old or too new")
	ErrAuth      = errors.New("event creator isn't team member")
)

type DagReader interface {
	GetEpoch() idx.Epoch
	GetMembers() pos.Members
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
	if e.Epoch != v.reader.GetEpoch() {
		return ErrNotRecent
	}
	if _, ok := v.reader.GetMembers()[e.Creator]; !ok {
		return ErrAuth
	}
	return nil
}
