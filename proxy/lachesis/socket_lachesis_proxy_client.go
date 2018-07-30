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
}

func NewSocketLachesisProxyClient(nodeAddr string, timeout time.Duration) *SocketLachesisProxyClient {
	return &SocketLachesisProxyClient{
		nodeAddr: nodeAddr,
		timeout:  timeout,
	}
}

func (p *SocketLachesisProxyClient) getConnection() (*rpc.Client, error) {
	conn, err := net.DialTimeout("tcp", p.nodeAddr, p.timeout)
	if err != nil {
		return nil, err
	}
	return jsonrpc.NewClient(conn), nil
}

func (p *SocketLachesisProxyClient) SubmitTx(tx []byte) (*bool, error) {
	rpcConn, err := p.getConnection()
	if err != nil {
		return nil, err
	}
	var ack bool
	err = rpcConn.Call("Lachesis.SubmitTx", tx, &ack)
	if err != nil {
		return nil, err
	}
	return &ack, nil
}
