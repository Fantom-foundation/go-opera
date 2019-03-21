package network

import (
	"context"
	"net"

	"google.golang.org/grpc"
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

// TcpConnect returns TCP grpc connection.
func TcpConnect(addr string) (*grpc.ClientConn, error) {
	return grpc.Dial(addr)
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

// FakeConnect returns fake grpc connection.
func FakeConnect(addr string) (*grpc.ClientConn, error) {
	return grpc.Dial(addr, grpc.WithContextDialer(fakeDial))
}

func fakeDial(ctx context.Context, addr string) (net.Conn, error) {
	return &Conn{
		remoteAddr: Addr(addr),
	}, nil
}
