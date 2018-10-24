package app

import (
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
	"sync"
	"time"

	"github.com/andrecronje/lachesis/src/poset"
	"github.com/andrecronje/lachesis/src/proxy/birpc"
	"github.com/andrecronje/lachesis/src/proxy/proto"
	ws "github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

type WebsocketAppProxy struct {
	//Save conn and not clients
	conn      map[*birpc.Connector]struct{}
	connMu    sync.Mutex
	rpcServer *rpc.Server

	clients  map[*rpc.Client]struct{} // TODO: remove clients on disconnect. websocket ping-pong?
	clientMu sync.Mutex

	submitCh chan []byte

	timeout time.Duration
	logger  *logrus.Logger
}

func NewWebsocketAppProxy(bindAddr string, timeout time.Duration, logger *logrus.Logger) (*WebsocketAppProxy, error) {
	if logger == nil {
		logger = logrus.New()
		logger.Level = logrus.DebugLevel
	}

	logger.WithFields(logrus.Fields{
		"bindAddr":     bindAddr,
		"timeout":      timeout,
	}).Debug("NewWebsocketAppProxy")



	proxy := WebsocketAppProxy{
		conn:     make(map[*birpc.Connector]struct{}),
		clients:  make(map[*rpc.Client]struct{}),
		submitCh: make(chan []byte),
		timeout:  timeout,
		logger:   logger,
	}

	go http.ListenAndServe(bindAddr, http.HandlerFunc(proxy.listen))

	return &proxy, nil
}

func (p *WebsocketAppProxy) listen(w http.ResponseWriter, r *http.Request) {
	upgrader := ws.Upgrader{}

	p.logger.Debug("func (p *WebsocketAppProxy) listen")

	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to Upgrade", http.StatusInternalServerError)
		p.logger.WithField("error", err).Error("Failed to Upgrade")
		return
	}

	// setup rpc
	conn := birpc.New(c, p.logger)
	//p.addClient(jsonrpc.NewClient(&p.conn.Client))
	p.addConn(conn)

	rpcServer := rpc.NewServer()
	rpcServer.RegisterName("Lachesis", p)

	p.logger.Debug("go p.rpcServer.ServeCodec(jsonrpc.NewServerCodec(&p.conn.Server))")
	go rpcServer.ServeCodec(jsonrpc.NewServerCodec(&conn.Server))
}

func (p *WebsocketAppProxy) addClient(c *rpc.Client) {
	p.clientMu.Lock()
	p.clients[c] = struct{}{}
	p.clientMu.Unlock()
}
func (p *WebsocketAppProxy) addConn(c *birpc.Connector) {
	p.connMu.Lock()
	p.conn[c] = struct{}{}
	p.connMu.Unlock()
}

func (p *WebsocketAppProxy) SubmitTx(tx []byte, ack *bool) error {
	p.logger.Debug("SubmitTx(tx []byte, ack *bool)")
	p.submitCh <- tx

	*ack = true

	return nil
}

//++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
//Implement AppProxy Interface

func (p *WebsocketAppProxy) SubmitCh() chan []byte {
	return p.submitCh
}

func (p *WebsocketAppProxy) CommitBlock(block poset.Block) ([]byte, error) {
	p.logger.Debug("CommitBlock(block poset.Block)")
	var stateHash proto.StateHash

	p.clientMu.Lock()
	defer p.clientMu.Unlock()

	p.logger.WithField("Clients", len(p.clients)).Debug("p.clients")
	for c, _ := range p.conn {
		if err := jsonrpc.NewClient(&c.Client).Call("State.CommitBlock", block, &stateHash); err != nil {
			p.logger.WithError(err).Debug("c.Call(, block, &stateHash)")
			return []byte{}, err
		}
		//Don't do an RPC call here, need to send to WS writer output


		/*if err := c.Call("State.CommitBlock", block, &stateHash); err != nil {
			p.logger.WithError(err).Debug("c.Call(, block, &stateHash)")
			return []byte{}, err
		}*/
	}

	p.logger.WithFields(logrus.Fields{
		"block":      block.Index(),
		"state_hash": stateHash.Hash,
	}).Debug("AppProxyClient.CommitBlock")

	return stateHash.Hash, nil // TODO: what to do with all the statehashes returned from each vm?
}

// TODO: move to vm side
func (p *WebsocketAppProxy) GetSnapshot(blockIndex int) ([]byte, error) {
	var snapshot proto.Snapshot

	p.clientMu.Lock()
	defer p.clientMu.Unlock()

	for c := range p.clients {
		if err := c.Call("State.GetSnapshot", blockIndex, &snapshot); err != nil {
			return []byte{}, err
		}
	}

	p.logger.WithFields(logrus.Fields{
		"block":    blockIndex,
		"snapshot": snapshot.Bytes,
	}).Debug("AppProxyClient.GetSnapshot")

	return snapshot.Bytes, nil
}

// TODO: move to vm side
func (p *WebsocketAppProxy) Restore(snapshot []byte) error {
	var stateHash proto.StateHash

	p.clientMu.Lock()
	defer p.clientMu.Unlock()

	for c := range p.clients {
		if err := c.Call("State.Restore", snapshot, &stateHash); err != nil {
			return err
		}
	}

	p.logger.WithFields(logrus.Fields{
		"state_hash": stateHash.Hash,
	}).Debug("AppProxyClient.Restore")

	return nil
}
