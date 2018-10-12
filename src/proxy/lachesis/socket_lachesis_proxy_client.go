package lachesis

import (
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"time"
)

type SocketLachesisProxyClient struct {
	nodeAddr string
	timeout  time.Duration
	rpc      *rpc.Client
}

func NewSocketLachesisProxyClient(nodeAddr string, timeout time.Duration) *SocketLachesisProxyClient {
	return &SocketLachesisProxyClient{
		nodeAddr: nodeAddr,
		timeout:  timeout,
	}
}

func (p *SocketLachesisProxyClient) getConnection() error {
	if p.rpc == nil {
		conn, err := net.DialTimeout("tcp", p.nodeAddr, p.timeout)

		if err != nil {
			return err
		}

		p.rpc = jsonrpc.NewClient(conn)
	}

	return nil
}

func (p *SocketLachesisProxyClient) SubmitTx(tx []byte) (*bool, error) {
	if err := p.getConnection(); err != nil {
		return nil, err
	}

	var ack bool

	err := p.rpc.Call("Lachesis.SubmitTx", tx, &ack)

	if err != nil {
		return nil, err
	}

	return &ack, nil
}

func (p *SocketLachesisProxyClient) Close() error {
	return p.rpc.Close()
}
