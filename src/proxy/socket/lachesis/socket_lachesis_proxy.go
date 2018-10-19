package lachesis

import (
	"fmt"
	"time"

	"github.com/andrecronje/lachesis/src/log"
	"github.com/andrecronje/lachesis/src/proxy"

	"github.com/sirupsen/logrus"
)

type SocketLachesisProxy struct {
	nodeAddress string
	bindAddress string

	handler proxy.ProxyHandler

	client *SocketLachesisProxyClientWebsocket
	server *SocketLachesisProxyServer
}

func NewSocketLachesisProxy(nodeAddr string,
	bindAddr string,
	handler proxy.ProxyHandler,
	timeout time.Duration,
	logger *logrus.Logger) (*SocketLachesisProxy, error) {

	if logger == nil {
		logger = logrus.New()
		logger.Level = logrus.DebugLevel
		lachesis_log.NewLocal(logger, logger.Level.String())
	}

	client := NewSocketLachesisProxyClientWebsocket(nodeAddr, timeout)

	server, err := NewSocketLachesisProxyServer(bindAddr, timeout, logger)

	if err != nil {
		return nil, err
	}

	proxy := &SocketLachesisProxy{
		nodeAddress: nodeAddr,
		bindAddress: bindAddr,
		handler:     handler,
		client:      client,
		server:      server,
	}

	go proxy.server.listen()

	return proxy, nil
}

func (p *SocketLachesisProxy) SubmitTx(tx []byte) error {
	err := p.client.SubmitTx(tx)
	if err != nil {
		return fmt.Errorf("Failed to deliver transaction to Lachesis: %v", err)
	}

	return nil
}
