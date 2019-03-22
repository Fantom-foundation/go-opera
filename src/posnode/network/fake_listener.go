package network

import (
	"io"
	"net"
)

// Listener is a fake net listener.
type Listener struct {
	NetAddr     Addr
	connections chan net.Conn
}

func NewListener(addr Addr) *Listener {
	return &Listener{
		NetAddr:     addr,
		connections: make(chan net.Conn, 10),
	}
}

func (l *Listener) connect() (net.Conn, error) {
	in1, out1 := io.Pipe()
	in2, out2 := io.Pipe()

	conn1 := &Conn{
		remoteAddr: Addr(":undefined"),
		input:      in1,
		output:     out2,
	}

	conn2 := &Conn{
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
	conn, got := <-l.connections
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
	close(l.connections)
	return nil
}

// Addr returns the listener's network address.
func (l *Listener) Addr() net.Addr {
	return l.NetAddr
}
