package peer

import "github.com/pkg/errors"

// Errors.
var (
	ErrTransportStopped      = errors.New("transport stopped")
	ErrClientProducerStopped = errors.New("client producer stopped")
	ErrReceiverIsBusy        = errors.New("receiver is busy")
	ErrProcessingTimeout     = errors.New("processing timeout")
	ErrBadResult             = errors.New("bad result")
	ErrServerAlreadyRunning  = errors.New("server already running")
)
