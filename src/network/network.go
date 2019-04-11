package network

import (
	"context"
	"net"
)

// ListenFunc returns addr listener.
type ListenFunc func(addr string) net.Listener

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
	res, err := listenFreeAddr(Addr(addr))
	if err != nil {
		panic(err)
	}

	return res
}

// FakeDialer returns fake connection creator.
func FakeDialer(from string) func(context.Context, string) (net.Conn, error) {
	return func(_ context.Context, addr string) (net.Conn, error) {
		listener := findListener(Addr(addr))
		if listener == nil {
			return nil, &net.AddrError{
				Err:  "connection refused",
				Addr: addr,
			}
		}

		return listener.connect(from)
	}
}
