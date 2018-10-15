package lachesis

import (
	"net/url"
	"time"

	ws "github.com/gorilla/websocket"
)

type SocketLachesisProxyClientWebsocket struct {
	nodeUrl string
	conn    *ws.Conn

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
		dialer := ws.DefaultDialer
		dialer.HandshakeTimeout = p.timeout

		c, _, err := dialer.Dial(p.nodeUrl, nil)
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

	p.conn.SetWriteDeadline(time.Now().Add(p.timeout))

	err := p.conn.WriteMessage(ws.BinaryMessage, tx)
	if err != nil {
		return err
	}

	return nil
}

func (p *SocketLachesisProxyClientWebsocket) Close() error {
	if p.conn == nil {
		return nil
	}

	err := p.conn.WriteControl(ws.CloseMessage, ws.FormatCloseMessage(ws.CloseNormalClosure, ""), time.Now().Add(p.timeout))
	if err != nil {
		p.conn.Close()
		return err
	}

	return p.conn.Close()
}
