package network

import (
	"net"
)

// Listener is a fake net listener.
type Listener struct {
	NetAddr Addr
}

/*
 * net.Listener implementation:
 */

// Accept waits for and returns the next connection to the listener.
func (l *Listener) Accept() (net.Conn, error) {
	return nil, nil
}

// Close closes the listener.
// Any blocked Accept operations will be unblocked and return errors.
func (l *Listener) Close() error {
	return nil
}

// Addr returns the listener's network address.
func (l *Listener) Addr() net.Addr {
	return Addr("fake_addr")
}
