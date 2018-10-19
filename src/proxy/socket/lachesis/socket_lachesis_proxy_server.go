package lachesis

import (
	"fmt"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"time"

	"github.com/andrecronje/lachesis/src/poset"
	"github.com/andrecronje/lachesis/src/proxy/proto"
	"github.com/sirupsen/logrus"
)

type SocketLachesisProxyServer struct {
	netListener       *net.Listener
	rpcServer         *rpc.Server
	commitCh          chan proto.Commit
	snapshotRequestCh chan proto.SnapshotRequest
	restoreCh         chan proto.RestoreRequest
	timeout           time.Duration
	logger            *logrus.Logger
}

func NewSocketLachesisProxyServer(bindAddress string,
	timeout time.Duration,
	logger *logrus.Logger) (*SocketLachesisProxyServer, error) {

	server := &SocketLachesisProxyServer{
		commitCh:          make(chan proto.Commit),
		snapshotRequestCh: make(chan proto.SnapshotRequest),
		restoreCh:         make(chan proto.RestoreRequest),
		timeout:           timeout,
		logger:            logger,
	}

	if err := server.register(bindAddress); err != nil {
		return nil, err
	}

	return server, nil
}

func (p *SocketLachesisProxyServer) register(bindAddress string) error {
	rpcServer := rpc.NewServer()
	rpcServer.RegisterName("State", p)
	p.rpcServer = rpcServer

	l, err := net.Listen("tcp", bindAddress)

	if err != nil {
		return err
	}

	p.netListener = &l

	return nil
}

func (p *SocketLachesisProxyServer) listen() error {
	for {
		conn, err := (*p.netListener).Accept()

		if err != nil {
			return err
		}

		go (*p.rpcServer).ServeCodec(jsonrpc.NewServerCodec(conn))
	}
}

func (p *SocketLachesisProxyServer) CommitBlock(block poset.Block, stateHash *proto.StateHash) (err error) {
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

func (p *SocketLachesisProxyServer) GetSnapshot(blockIndex int, snapshot *proto.Snapshot) (err error) {
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

func (p *SocketLachesisProxyServer) Restore(snapshot []byte, stateHash *proto.StateHash) (err error) {
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
