package ordering

import (
	"fmt"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"

	"github.com/Fantom-foundation/go-lachesis/src/inter"
)

type lamportTimeValidator struct {
	event   *inter.Event
	maxTime idx.Lamport
}

func newLamportTimeValidator(e *inter.Event) *lamportTimeValidator {
	return &lamportTimeValidator{
		event:   e,
		maxTime: 0,
	}
}

func (v *lamportTimeValidator) AddParentTime(time idx.Lamport) error {
	if v.event.Lamport <= time {
		return fmt.Errorf("event %s has lamport time %d. It isn't next of parents",
			v.event.Hash().String(),
			v.event.Lamport)
	}
	if v.maxTime < time {
		v.maxTime = time
	}
	return nil
}

func (v *lamportTimeValidator) CheckSequential() error {
	if v.event.Lamport != v.maxTime+1 {
		return fmt.Errorf("event %s has lamport time %d. It is too far from parents",
			v.event.Hash().String(),
			v.event.Lamport)
	}
	return nil
}
