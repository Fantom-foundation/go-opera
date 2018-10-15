package lachesis

import (
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

type SocketLachesisProxyClientWebsocket struct {
	nodeUrl string
	conn    *websocket.Conn

	timeout time.Duration
}

func NewSocketLachesisProxyClientWebsocket(nodeAddr string, timeout time.Duration) *SocketLachesisProxyClientWebsocket {
	u := url.URL{Scheme: "ws", Host: nodeAddr}

	return &SocketLachesisProxyClientWebsocket{
		nodeUrl: u.String(),
		timeout: timeout,
	}
}

func (p *SocketLachesisProxyClientWebsocket) getConnection() error {
	if p.conn == nil {
		c, _, err := websocket.DefaultDialer.Dial(p.nodeUrl, nil)
		if err != nil {
			return err
		}

		p.conn = c
	}

	return nil
}

func (p *SocketLachesisProxyClientWebsocket) SubmitTx(tx []byte) error {
	if err := p.getConnection(); err != nil {
		return err
	}

	err := p.conn.WriteMessage(websocket.BinaryMessage, tx)
	if err != nil {
		return err
	}

	return nil
}

func (p *SocketLachesisProxyClientWebsocket) Close() error {
	if p.conn != nil {
		return p.conn.Close()
	}
	return nil
}
