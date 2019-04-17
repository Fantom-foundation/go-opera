package network

import (
	"fmt"
	"io"
	"net"
	"sync"
)

// Listener is a fake net listener.
type Listener struct {
	NetAddr     Addr
	connections chan net.Conn
	sync.RWMutex
}

// NewListener makes fake listener.
func NewListener(addr Addr) *Listener {
	return &Listener{
		NetAddr:     addr,
		connections: make(chan net.Conn, 10),
	}
}

func (l *Listener) connect(from string) (net.Conn, error) {
	from = from + ":0"

	l.RLock()
	defer l.RUnlock()

	if l.connections == nil {
		return nil, &net.OpError{
			Op:     "connect",
			Net:    "fake",
			Source: Addr(from),
			Addr:   l.Addr(),
			Err:    fmt.Errorf("closed"),
		}
	}

	in1, out1 := io.Pipe()
	in2, out2 := io.Pipe()

	conn1 := &Conn{
		localAddr:  l.NetAddr,
		remoteAddr: Addr(from),
		input:      in1,
		output:     out2,
	}

	conn2 := &Conn{
		localAddr:  Addr(from),
		remoteAddr: l.NetAddr,
		input:      in2,
		output:     out1,
	}

	l.connections <- conn1
	return conn2, nil
}

/*
 * net.Listener implementation:
 */

// Accept waits for and returns the next connection to the listener.
func (l *Listener) Accept() (net.Conn, error) {
	l.RLock()
	connections := l.connections
	l.RUnlock()

	if connections == nil {
		return nil, &net.OpError{
			Op:     "connect",
			Net:    "fake",
			Source: l.Addr(),
			Err:    fmt.Errorf("closed"),
		}
	}

	conn, got := <-connections
	if !got {
		return nil, &net.AddrError{
			Err:  "listener closed",
			Addr: l.Addr().String(),
		}
	}
	return conn, nil
}

// Close closes the listener.
// Any blocked Accept operations will be unblocked and return errors.
func (l *Listener) Close() error {
	removeListener(l)

	l.Lock()
	defer l.Unlock()

	if l.connections == nil {
		return &net.OpError{
			Op:     "close",
			Net:    "fake",
			Source: l.Addr(),
			Err:    fmt.Errorf("closed already"),
		}
	}

	close(l.connections)
	l.connections = nil
	return nil
}

// Addr returns the listener's network address.
func (l *Listener) Addr() net.Addr {
	return l.NetAddr
}
