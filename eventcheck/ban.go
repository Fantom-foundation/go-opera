package eventcheck

import (
	base "github.com/Fantom-foundation/lachesis-base/eventcheck"
	"github.com/Fantom-foundation/lachesis-base/eventcheck/epochcheck"
)

var (
	ErrAlreadyConnectedEvent = base.ErrAlreadyConnectedEvent
)

func IsBan(err error) bool {
	if err == epochcheck.ErrNotRelevant ||
		err == ErrAlreadyConnectedEvent {
		return false
	}
	return err != nil
}
