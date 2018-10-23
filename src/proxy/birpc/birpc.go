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
)

type rpcMessage struct {
	Method *json.RawMessage `json:"method"`
	Result *json.RawMessage `json:"result"`
}

type rpcConnector struct {
	w *io.PipeWriter
	r *io.PipeReader

	err     error
	errLock sync.Mutex
}

func (c *rpcConnector) Read(p []byte) (int, error) {
	/*if c.error() != nil {
		return 0, c.error()
	}*/
	return c.r.Read(p)
}

func (c *rpcConnector) Write(p []byte) (int, error) {
	/*if c.error() != nil {
		return 0, c.error()
	}*/
	return c.w.Write(p)
}

func (c *rpcConnector) Close() error {
	c.r.Close()
	c.w.Close()
	return nil
}

func (c *rpcConnector) error() (err error) {
	c.errLock.Lock()
	err = c.err
	c.errLock.Unlock()
	return
}

func (c *rpcConnector) setError(err error) {
	c.errLock.Lock()
	if c.err != nil && err != nil {
		c.err = err
	}
	c.errLock.Unlock()
}

type Connector struct {
	Server rpcConnector
	Client rpcConnector

	conn *ws.Conn
}

func New(ws *ws.Conn) *Connector {
	wsrw := Connector{
		conn: ws,
	}

	wsrw.initWsReader()
	wsrw.initWsWriter()
	return &wsrw
}

func (x *Connector) Close() error {
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
		for ; x.Server.error() == nil && err == nil; _, b, err = x.conn.ReadMessage() {
			var rpcMsg rpcMessage
			err = json.Unmarshal(b, &rpcMsg)
			if err != nil {
				break
			}
			if rpcMsg.Method != nil {
				// must go to rpc server
				_, err = sw.Write(b)
			} else if rpcMsg.Result != nil {
				// must go ro rpc client
				_, err = cw.Write(b)
			}
		}
		x.Server.setError(err)
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
		for ; x.Server.error() == nil && err == nil; err = dec.Decode(&b) {
			wsWriteLock.Lock()
			err = x.conn.WriteMessage(ws.TextMessage, b)
			wsWriteLock.Unlock()
		}
		x.Server.setError(err)
		sr.Close()
	}(x)
	go func(x *Connector) {
		dec := json.NewDecoder(cr)
		var b json.RawMessage
		err := dec.Decode(&b)
		for ; x.Client.error() == nil && err == nil; err = dec.Decode(&b) {
			wsWriteLock.Lock()
			err = x.conn.WriteMessage(ws.TextMessage, b)
			wsWriteLock.Unlock()
		}
		x.Client.setError(err)
		cr.Close()
	}(x)
}
