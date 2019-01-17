package fakenet

import (
	"io"
	"net"
	"time"
)

// Conn is a fake connection.
type Conn struct {
	LAddress string
	RAddress string
	Reader   *io.PipeReader
	Writer   *io.PipeWriter
}

// Read reads data from the fake connection.
func (c *Conn) Read(b []byte) (int, error) {
	return c.Reader.Read(b)
}

// Write writes data to the fake connection.
func (c *Conn) Write(b []byte) (int, error) {
	return c.Writer.Write(b)
}

// Close closes the fake connection.
func (c *Conn) Close() error {
	err := c.Writer.Close()
	if err == nil {
		err = c.Reader.Close()
	}
	return err
}

// LocalAddr returns the local network address.
func (c *Conn) LocalAddr() net.Addr {
	return Addr{
		AddressString: c.LAddress,
		NetworkString: "tcp",
	}
}

// RemoteAddr returns the local network address.
func (c *Conn) RemoteAddr() net.Addr {
	return Addr{
		AddressString: c.RAddress,
		NetworkString: "tcp",
	}
}

// SetDeadline sets the read and write deadlines associated
// with the connection. Not implemented.
func (c *Conn) SetDeadline(t time.Time) error { return nil }

// SetReadDeadline sets the deadline for future Read calls
// and any currently-blocked Read call. Not implemented.
func (c *Conn) SetReadDeadline(t time.Time) error { return nil }

// SetWriteDeadline sets the deadline for future Write calls
// and any currently-blocked Write call. Not implemented.
func (c *Conn) SetWriteDeadline(t time.Time) error { return nil }
