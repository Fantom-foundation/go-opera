package lachesis

import (
	"fmt"
	"net/rpc"
	"net/rpc/jsonrpc"
	"net/url"
	"time"

	"github.com/andrecronje/lachesis/src/poset"
	"github.com/andrecronje/lachesis/src/proxy"
	"github.com/andrecronje/lachesis/src/proxy/birpc"
	"github.com/andrecronje/lachesis/src/proxy/proto"
	ws "github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

type WebsocketLachesisProxy struct {
	conn      *birpc.Connector
	rpcServer *rpc.Server
	rpcClient *rpc.Client
	handler   proxy.ProxyHandler // ?

	commitCh          chan proto.Commit
	snapshotRequestCh chan proto.SnapshotRequest
	restoreCh         chan proto.RestoreRequest

	nodeUrl string
	timeout time.Duration
	logger  *logrus.Logger
}

func NewWebsocketLachesisProxy(nodeAddr string,
	handler proxy.ProxyHandler,
	timeout time.Duration,
	logger *logrus.Logger) (*WebsocketLachesisProxy, error) {

	u := url.URL{Scheme: "ws", Host: nodeAddr}
	if logger == nil {
		logger = logrus.New()
		logger.Level = logrus.DebugLevel
	}

	proxy := WebsocketLachesisProxy{
		handler:           handler,
		commitCh:          make(chan proto.Commit),
		snapshotRequestCh: make(chan proto.SnapshotRequest),
		restoreCh:         make(chan proto.RestoreRequest),
		nodeUrl:           u.String(),
		timeout:           timeout,
		logger:            logger,
	}

	if err := proxy.getConnection(); err != nil {
		return nil, err
	}

	return &proxy, nil
}

func (p *WebsocketLachesisProxy) getConnection() error {
	if p.conn != nil {
		return nil
	}

	dialer := ws.DefaultDialer
	dialer.HandshakeTimeout = p.timeout

	c, _, err := dialer.Dial(p.nodeUrl, nil)
	if err != nil {
		return err
	}

	// setup rpc
	p.conn = birpc.New(c)
	rpcServer := rpc.NewServer()
	rpcServer.RegisterName("State", p)
	p.rpcServer = rpcServer
	p.rpcClient = jsonrpc.NewClient(&p.conn.Client)
	go p.rpcServer.ServeCodec(jsonrpc.NewServerCodec(&p.conn.Server))
	return nil
}

func (p *WebsocketLachesisProxy) CommitBlock(block poset.Block, stateHash *proto.StateHash) (err error) {
	// Send the Commit over
	respCh := make(chan proto.CommitResponse)

	p.commitCh <- proto.Commit{
		Block:    block,
		RespChan: respCh,
	}

	// Wait for a response
	select {
	case commitResp := <-respCh:
		stateHash.Hash = commitResp.StateHash

		if commitResp.Error != nil {
			err = commitResp.Error
		}

	case <-time.After(p.timeout):
		err = fmt.Errorf("command timed out")
	}

	p.logger.WithFields(logrus.Fields{
		"block":      block.Index(),
		"state_hash": stateHash.Hash,
		"err":        err,
	}).Debug("LachesisProxyServer.CommitBlock")

	return
}

func (p *WebsocketLachesisProxy) GetSnapshot(blockIndex int, snapshot *proto.Snapshot) (err error) {
	// Send the Request over
	respCh := make(chan proto.SnapshotResponse)

	p.snapshotRequestCh <- proto.SnapshotRequest{
		BlockIndex: blockIndex,
		RespChan:   respCh,
	}

	// Wait for a response
	select {
	case snapshotResp := <-respCh:
		snapshot.Bytes = snapshotResp.Snapshot

		if snapshotResp.Error != nil {
			err = snapshotResp.Error
		}

	case <-time.After(p.timeout):
		err = fmt.Errorf("command timed out")
	}

	p.logger.WithFields(logrus.Fields{
		"block":    blockIndex,
		"snapshot": snapshot.Bytes,
		"err":      err,
	}).Debug("LachesisProxyServer.GetSnapshot")

	return
}

func (p *WebsocketLachesisProxy) Restore(snapshot []byte, stateHash *proto.StateHash) (err error) {
	// Send the Request over
	respCh := make(chan proto.RestoreResponse)

	p.restoreCh <- proto.RestoreRequest{
		Snapshot: snapshot,
		RespChan: respCh,
	}

	// Wait for a response
	select {
	case restoreResp := <-respCh:
		stateHash.Hash = restoreResp.StateHash

		if restoreResp.Error != nil {
			err = restoreResp.Error
		}

	case <-time.After(p.timeout):
		err = fmt.Errorf("command timed out")
	}

	p.logger.WithFields(logrus.Fields{
		"state_hash": stateHash.Hash,
		"err":        err,
	}).Debug("LachesisProxyServer.Restore")

	return
}

func (p *WebsocketLachesisProxy) Close() error {
	p.conn.Close()
	return p.rpcClient.Close()
}

//++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
//Implement LachesisProxy interface

func (p *WebsocketLachesisProxy) CommitCh() chan proto.Commit {
	return p.commitCh
}

func (p *WebsocketLachesisProxy) SnapshotRequestCh() chan proto.SnapshotRequest {
	return p.snapshotRequestCh
}

func (p *WebsocketLachesisProxy) RestoreCh() chan proto.RestoreRequest {
	return p.restoreCh
}

func (p *WebsocketLachesisProxy) SubmitTx(tx []byte) error {
	errMsg := func(err string) error {
		return fmt.Errorf("failed to deliver transaction to Lachesis: %s", err)
	}
	if err := p.getConnection(); err != nil {
		return errMsg(err.Error())
	}

	var ack bool

	err := p.rpcClient.Call("Lachesis.SubmitTx", tx, &ack)
	if err != nil {
		return errMsg(err.Error())
	}

	if !ack {
		errMsg("no ack")
	}

	return nil
}
