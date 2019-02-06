package fakenet

import "github.com/pkg/errors"

// Errors.
var (
	ErrListenerClosed      = errors.New("listener closed")
	ErrAddressAlreadyInUse = errors.New("address already in use")
)
