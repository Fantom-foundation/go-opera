package peer

import (
	"context"
	"sync"

	"github.com/sirupsen/logrus"
)

// NewSyncClientFunc creates a new sync client.
type NewSyncClientFunc func(target string) (SyncClient, error)

// SyncPeer is an interface representing methods for sync transport.
type SyncPeer interface {
	Sync(ctx context.Context, target string,
		req *SyncRequest, resp *SyncResponse) error
	ForceSync(ctx context.Context, target string,
		req *ForceSyncRequest, resp *ForceSyncResponse) error
	FastForward(ctx context.Context, target string,
		req *FastForwardRequest, resp *FastForwardResponse) error
	ReceiverChannel() <-chan *RPC
	Close() error
}

// Peer implements SyncPeer interface.
type Peer struct {
	clientProducer ClientProducer
	logger         logrus.FieldLogger
	server         SyncServer

	mtx      sync.RWMutex
	shutdown bool

	wg *sync.WaitGroup
}

// NewTransport creates a net transport.
func NewTransport(logger logrus.FieldLogger,
	clientProducer ClientProducer, server SyncServer) *Peer {
	logger = logger.WithField("type", "transport")
	return &Peer{
		clientProducer: clientProducer,
		logger:         logger,
		server:         server,
		wg:             &sync.WaitGroup{},
	}
}

// Sync creates a sync request to a specific node.
func (tr *Peer) Sync(ctx context.Context, target string,
	req *SyncRequest, resp *SyncResponse) error {
	if tr.isShutdown() {
		return ErrTransportStopped
	}

	tr.wg.Add(1)
	defer tr.wg.Done()

	return tr.sync(ctx, target, req, resp)
}

func (tr *Peer) sync(ctx context.Context, target string,
	req *SyncRequest, resp *SyncResponse) error {
	logger := tr.logger.WithFields(logrus.Fields{"method": "sync",
		"target": target})

	cli, err := tr.clientProducer.Pop(target)
	if err != nil {
		logger.Error(err)
		return err
	}

	if err := cli.Sync(ctx, req, resp); err != nil {
		logger.Error(err)
		return err
	}
	tr.clientProducer.Push(target, cli)

	return err
}

// ForceSync creates a force sync request to a specific node.
func (tr *Peer) ForceSync(ctx context.Context, target string,
	req *ForceSyncRequest, resp *ForceSyncResponse) error {
	if tr.isShutdown() {
		return ErrTransportStopped
	}

	tr.wg.Add(1)
	defer tr.wg.Done()

	return tr.forceSync(ctx, target, req, resp)
}

func (tr *Peer) forceSync(ctx context.Context, target string,
	req *ForceSyncRequest, resp *ForceSyncResponse) error {
	logger := tr.logger.WithFields(logrus.Fields{"method": "forceSync",
		"target": target})

	cli, err := tr.clientProducer.Pop(target)
	if err != nil {
		logger.Error(err)
		return err
	}

	if err := cli.ForceSync(ctx, req, resp); err != nil {
		logger.Error(err)
		return err
	}
	tr.clientProducer.Push(target, cli)

	return nil
}

// FastForward creates a fast forward request to a specific node.
func (tr *Peer) FastForward(ctx context.Context, target string,
	req *FastForwardRequest, resp *FastForwardResponse) error {

	if tr.isShutdown() {
		return ErrTransportStopped
	}

	tr.wg.Add(1)
	defer tr.wg.Done()

	return tr.fastForward(ctx, target, req, resp)
}

func (tr *Peer) fastForward(ctx context.Context, target string,
	req *FastForwardRequest, resp *FastForwardResponse) error {
	logger := tr.logger.WithFields(logrus.Fields{"method": "fastForward",
		"target": target})

	cli, err := tr.clientProducer.Pop(target)
	if err != nil {
		logger.Error(err)
		return err
	}

	if err := cli.FastForward(ctx, req, resp); err != nil {
		logger.Error(err)
		return err
	}
	tr.clientProducer.Push(target, cli)

	return nil
}

// ReceiverChannel returns a sync server receiver channel.
func (tr *Peer) ReceiverChannel() <-chan *RPC {
	tr.mtx.Lock()
	defer tr.mtx.Unlock()

	return tr.server.ReceiverChannel()
}

// Close closes the transport.
func (tr *Peer) Close() error {
	logger := tr.logger.WithField("method", "Close")

	tr.mtx.Lock()
	defer tr.mtx.Unlock()

	if tr.shutdown {
		return nil
	}
	tr.shutdown = true

	// Stop accepting new connections.
	if tr.server != nil {
		if err := tr.server.Close(); err != nil {
			logger.Error(err)
		}
	}

	// Stop creating new connections.
	tr.clientProducer.Close()

	// Waiting for all outgoing connections to complete.
	tr.wg.Wait()

	return nil
}

func (tr *Peer) isShutdown() bool {
	tr.mtx.Lock()
	defer tr.mtx.Unlock()
	return tr.shutdown
}
