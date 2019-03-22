package network

import (
	"context"
	"net"
)

// TcpListener returns TCP listener binded to addr.
// Leave addr empty to get any free addr.
func TcpListener(addr string) net.Listener {
	res, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}
	return res
}

// FakeListener returns fake listener binded to addr.
// Leave addr empty to get any free addr.
func FakeListener(addr string) net.Listener {
	if addr == "" {
		// TODO: get free addrs
		addr = "random addr"
	} else {
		// TODO: check addr free
	}
	return &Listener{
		NetAddr: Addr(addr),
	}
}

// FakeDial returns fake connection.
func FakeDial(ctx context.Context, addr string) (net.Conn, error) {
	return &Conn{
		remoteAddr: Addr(addr),
	}, nil
}
