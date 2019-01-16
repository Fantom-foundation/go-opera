package fakenet

import (
	"net"
	"sync"
)

// Listener is a fake network listener.
type Listener struct {
	done    chan struct{}
	Address string
	Input   chan *Conn

	mtx      sync.RWMutex
	shutdown bool
}

// NewListener creates a new fake listener.
func NewListener(address string) *Listener {
	return &Listener{
		done:    make(chan struct{}),
		Address: address,
		Input:   make(chan *Conn),
	}
}

// Accept waits for and returns the next connection to the fake listener.
func (l *Listener) Accept() (conn net.Conn, err error) {
	select {
	case conn = <-l.Input:
		return conn, nil
	case <-l.done:
		return nil, ErrListenerClosed
	}
}

// Close closes the listener.
func (l *Listener) Close() error {
	l.mtx.Lock()
	defer l.mtx.Unlock()

	if l.shutdown {
		return nil
	}
	l.shutdown = true

	close(l.done)
	return nil
}

// Addr returns the listener's fake network address.
func (l *Listener) Addr() net.Addr {
	return Addr{
		AddressString: l.Address,
		NetworkString: "tcp",
	}
}

func (l *Listener) isClosed() bool {
	l.mtx.RLock()
	defer l.mtx.RUnlock()
	return l.shutdown
}
