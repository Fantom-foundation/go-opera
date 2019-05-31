package ordering

import (
	"fmt"

	"github.com/Fantom-foundation/go-lachesis/src/inter"
)

type lamportTimeValidator struct {
	event   *inter.Event
	maxTime inter.Timestamp
}

func newLamportTimeValidator(e *inter.Event) *lamportTimeValidator {
	return &lamportTimeValidator{
		event:   e,
		maxTime: 0,
	}
}

func (v *lamportTimeValidator) AddParentTime(time inter.Timestamp) error {
	if v.event.LamportTime <= time {
		return fmt.Errorf("event %s has lamport time %d. It isn't next of parents",
			v.event.Hash().String(),
			v.event.LamportTime)
	}
	if v.maxTime < time {
		v.maxTime = time
	}
	return nil
}

func (v *lamportTimeValidator) CheckSequential() error {
	if v.event.LamportTime != v.maxTime+1 {
		return fmt.Errorf("event %s has lamport time %d. It is too far from parents",
			v.event.Hash().String(),
			v.event.LamportTime)
	}
	return nil
}
