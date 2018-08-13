package lachesis

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

type SocketLachesisProxy struct {
	nodeAddress string
	bindAddress string

	client *SocketLachesisProxyClient
	server *SocketLachesisProxyServer
}

func NewSocketLachesisProxy(nodeAddr string,
	bindAddr string,
	timeout time.Duration,
	logger *logrus.Logger) (*SocketLachesisProxy, error) {

	if logger == nil {
		logger = logrus.New()
		logger.Level = logrus.DebugLevel
	}

	client := NewSocketLachesisProxyClient(nodeAddr, timeout)
	server, err := NewSocketLachesisProxyServer(bindAddr, timeout, logger)
	if err != nil {
		return nil, err
	}

	proxy := &SocketLachesisProxy{
		nodeAddress: nodeAddr,
		bindAddress: bindAddr,
		client:      client,
		server:      server,
	}
	go proxy.server.listen()

	return proxy, nil
}

//++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
//Implement LachesisProxy interface

func (p *SocketLachesisProxy) CommitCh() chan Commit {
	return p.server.commitCh
}

func (p *SocketLachesisProxy) SubmitTx(tx []byte) error {
	ack, err := p.client.SubmitTx(tx)
	if err != nil {
		return err
	}
	if !*ack {
		return fmt.Errorf("failed to deliver transaction to Lachesis")
	}
	return nil
}
