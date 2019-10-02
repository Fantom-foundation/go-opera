package event_check

import (
	"errors"

	"github.com/Fantom-foundation/go-lachesis/src/event_check/epoch_check"
)

var (
	ErrAlreadyConnectedEvent = errors.New("event is connected already")
)

func IsBan(err error) bool {
	if err == epoch_check.ErrNotRelevant ||
		err == ErrAlreadyConnectedEvent {
		return false
	}
	return err != nil
}
