package fakenet

import "github.com/pkg/errors"

var (
	ErrListenerClosed = errors.New("listener closed")
)
