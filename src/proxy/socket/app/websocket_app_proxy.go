package app

import (
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
	"time"

	"github.com/andrecronje/lachesis/src/poset"
	"github.com/andrecronje/lachesis/src/proxy/birpc"
	"github.com/andrecronje/lachesis/src/proxy/proto"
	ws "github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

type WebsocketAppProxy struct {
	conn      *birpc.Connector
	rpcServer *rpc.Server
	rpcClient *rpc.Client
	// TODO: Handle more than one ws connection.
	// What do we do with mulitple EVM's?
	// Do we commit blocks to all or only one (randomly)?

	submitCh chan []byte

	timeout time.Duration
	logger  *logrus.Logger
}

func NewWebsocketAppProxy(bindAddr string, timeout time.Duration, logger *logrus.Logger) (*WebsocketAppProxy, error) {
	if logger == nil {
		logger = logrus.New()
		logger.Level = logrus.DebugLevel
	}

	proxy := WebsocketAppProxy{
		submitCh: make(chan []byte),
		timeout:  timeout,
		logger:   logger,
	}

	go http.ListenAndServe(bindAddr, http.HandlerFunc(proxy.listen))

	return &proxy, nil
}

func (p *WebsocketAppProxy) listen(w http.ResponseWriter, r *http.Request) {
	upgrader := ws.Upgrader{}

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to Upgrade", http.StatusInternalServerError)
		p.logger.WithField("error", err).Error("Failed to Upgrade")
		return
	}

	if p.conn != nil {
		return
	}

	// setup rpc
	p.conn = birpc.New(c)
	rpcServer := rpc.NewServer()
	rpcServer.RegisterName("Lachesis", p)
	p.rpcServer = rpcServer
	p.rpcClient = jsonrpc.NewClient(&p.conn.Client)
	go p.rpcServer.ServeCodec(jsonrpc.NewServerCodec(&p.conn.Server))
}

func (p *WebsocketAppProxy) SubmitTx(tx []byte, ack *bool) error {
	p.logger.Debug("SubmitTx")
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
	var stateHash proto.StateHash

	if err := p.rpcClient.Call("State.CommitBlock", block, &stateHash); err != nil {
		return []byte{}, err
	}

	p.logger.WithFields(logrus.Fields{
		"block":      block.Index(),
		"state_hash": stateHash.Hash,
	}).Debug("AppProxyClient.CommitBlock")

	return stateHash.Hash, nil
}

func (p *WebsocketAppProxy) GetSnapshot(blockIndex int) ([]byte, error) {
	var snapshot proto.Snapshot

	if err := p.rpcClient.Call("State.GetSnapshot", blockIndex, &snapshot); err != nil {
		return []byte{}, err
	}

	p.logger.WithFields(logrus.Fields{
		"block":    blockIndex,
		"snapshot": snapshot.Bytes,
	}).Debug("AppProxyClient.GetSnapshot")

	return snapshot.Bytes, nil
}

func (p *WebsocketAppProxy) Restore(snapshot []byte) error {
	var stateHash proto.StateHash

	if err := p.rpcClient.Call("State.Restore", snapshot, &stateHash); err != nil {
		return err
	}

	p.logger.WithFields(logrus.Fields{
		"state_hash": stateHash.Hash,
	}).Debug("AppProxyClient.Restore")

	return nil
}
