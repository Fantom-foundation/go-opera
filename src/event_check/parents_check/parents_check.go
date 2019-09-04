package parents_check

import (
	"errors"

	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/inter/pos"
	"github.com/Fantom-foundation/go-lachesis/src/lachesis"
)

var (
	ErrWrongSeq        = errors.New("event has wrong sequence time")
	ErrWrongLamport    = errors.New("event has wrong Lamport time")
	ErrDoubleParents   = errors.New("event has double parents")
	ErrWrongSelfParent = errors.New("event is missing self-parent")
)

type DagReader interface {
	GetMembers() pos.Members
}

// Check which require only parents list + current epoch info
type Validator struct {
	config *lachesis.DagConfig
	reader DagReader
}

// Performs checks, which require known the parents
func New(config *lachesis.DagConfig, reader DagReader) *Validator {
	return &Validator{
		config: config,
		reader: reader,
	}
}

func (v *Validator) validateGasLeft(e *inter.Event, parents []*inter.EventHeaderData) error {
	// TODO validate e.GasPowerLeft against median time, self-parent, and creator's stake
	return nil
}

func (v *Validator) Validate(e *inter.Event, parents []*inter.EventHeaderData) error {
	if len(e.Parents) != len(parents) {
		panic("parents_check: expected event's parents as an argument")
	}

	// lamport
	maxLamport := idx.Lamport(0)
	for _, p := range parents {
		maxLamport = idx.MaxLamport(maxLamport, p.Lamport)
	}
	if e.Lamport != maxLamport+1 {
		return ErrWrongLamport
	}

	// parents
	if len(e.Parents.Set()) != len(e.Parents) {
		return ErrDoubleParents
	}

	// self-parent
	for i, p := range parents {
		if (p.Creator == e.Creator) != e.IsSelfParent(e.Parents[i]) {
			return ErrWrongSelfParent
		}
	}

	// seq
	if (e.Seq <= 1) != (e.SelfParent() == nil) {
		return ErrWrongSeq
	}
	if e.SelfParent() != nil {
		selfParent := parents[0]
		if !e.IsSelfParent(selfParent.Hash()) {
			// sanity check, self-parent is always first, it's how it's stored
			return ErrWrongSelfParent
		}
		if e.Seq != selfParent.Seq+1 {
			return ErrWrongSeq
		}
	}

	// gas left
	if err := v.validateGasLeft(e, parents); err != nil {
		return err
	}

	return nil
}
