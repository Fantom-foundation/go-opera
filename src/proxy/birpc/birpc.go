// Bi-directional rpc.
// Allows both rpc server & client to run over single connection.
// Splits incoming traffic and forwards it to correct service.
// Combines outgoing traffic. (Needs bi-rpc at other end as well)

package birpc

import (
	"encoding/json"
	"io"
	"sync"

	ws "github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

type rpcMessage struct {
	Method *json.RawMessage `json:"method"`
	Result *json.RawMessage `json:"result"`
}

type rpcConnector struct {
	w     *io.PipeWriter
	r     *io.PipeReader
	error func() error
}

func (c *rpcConnector) Read(p []byte) (int, error) {
	if err := c.error(); err != nil {
		return 0, err
	}
	return c.r.Read(p)
}

func (c *rpcConnector) Write(p []byte) (int, error) {
	if err := c.error(); err != nil {
		return 0, err
	}
	return c.w.Write(p)
}

func (c *rpcConnector) Close() error {
	c.r.Close()
	c.w.Close()
	return nil
}

type Connector struct {
	Server rpcConnector
	Client rpcConnector
	conn   *ws.Conn

	err     error
	errLock sync.Mutex
	logger  *logrus.Logger
}

func New(ws *ws.Conn, logger *logrus.Logger) *Connector {
	wsrw := Connector{
		conn: ws,
		logger: logger,
	}
	wsrw.Server.error = wsrw.error
	wsrw.Client.error = wsrw.error

	wsrw.initWsReader()
	wsrw.initWsWriter()
	return &wsrw
}

func (x *Connector) Close() error {

	x.conn.Close()
	x.Server.Close()
	x.Client.Close()
	return nil
}

func (x *Connector) initWsReader() {
	var sw *io.PipeWriter
	var cw *io.PipeWriter
	x.Server.r, sw = io.Pipe()
	x.Client.r, cw = io.Pipe()

	go func(x *Connector) {
		_, b, err := x.conn.ReadMessage()
		x.logger.Debug("_, b, err := x.conn.ReadMessage()")
		for ; err == nil && x.error() == nil; _, b, err = x.conn.ReadMessage() {
			if err != nil {
				x.logger.WithError(err).Debug("x.conn.ReadMessage()")
			}
			var rpcMsg rpcMessage
			err = json.Unmarshal(b, &rpcMsg)
			if err != nil {
				x.logger.WithError(err).Debug("json.Unmarshal(b, &rpcMsg)")
				break
			}
			if rpcMsg.Method != nil {
				// must go to rpc server
				_, err = sw.Write(b)
				if err != nil {
					x.logger.WithError(err).Debug("sw.Write(b)")
				}
			} else if rpcMsg.Result != nil {
				// must go ro rpc client
				_, err = cw.Write(b)
				if err != nil {
					x.logger.WithError(err).Debug("cw.Write(b)")
				}
			}
		}
		x.setError(err)
		sw.Close()
		cw.Close()
	}(x)
}

func (x *Connector) initWsWriter() {
	var sr *io.PipeReader
	var cr *io.PipeReader
	sr, x.Server.w = io.Pipe()
	cr, x.Client.w = io.Pipe()
	wsWriteLock := sync.Mutex{}

	go func(x *Connector) {
		dec := json.NewDecoder(sr)
		var b json.RawMessage
		err := dec.Decode(&b)
		for ; err == nil && x.error() == nil; err = dec.Decode(&b) {
			wsWriteLock.Lock()
			err = x.conn.WriteMessage(ws.TextMessage, b)
			wsWriteLock.Unlock()
		}
		x.setError(err)
		sr.Close()
		cr.Close()
	}(x)
	go func(x *Connector) {
		dec := json.NewDecoder(cr)
		var b json.RawMessage
		err := dec.Decode(&b)
		for ; err == nil && x.error() == nil; err = dec.Decode(&b) {
			wsWriteLock.Lock()
			err = x.conn.WriteMessage(ws.TextMessage, b)
			wsWriteLock.Unlock()
		}
		x.setError(err)
		cr.Close()
		sr.Close()
	}(x)
}

func (x *Connector) error() (err error) {
	x.errLock.Lock()
	err = x.err
	x.errLock.Unlock()
	return
}

func (x *Connector) setError(err error) {
	x.errLock.Lock()
	if x.err != nil && err != nil {
		x.err = err
	}
	x.errLock.Unlock()
}
