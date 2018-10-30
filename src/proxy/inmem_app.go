package proxy

import (
	"github.com/sirupsen/logrus"

	"github.com/andrecronje/lachesis/src/poset"
)

// InmemAppProxy implements the AppProxy interface natively
type InmemAppProxy struct {
	logger   *logrus.Logger
	handler  ProxyHandler
	submitCh chan []byte
}

// NewInmemAppProxy instantiates an InmemProxy from a set of handlers
func NewInmemAppProxy(handler ProxyHandler, logger *logrus.Logger) *InmemAppProxy {
	if logger == nil {
		logger = logrus.New()
		logger.Level = logrus.DebugLevel
	}

	return &InmemAppProxy{
		logger:   logger,
		handler:  handler,
		submitCh: make(chan []byte),
	}
}

/*
 * inmem interface: AppProxy implementation
 */

// SubmitCh implements AppProxy interface method
func (p *InmemAppProxy) SubmitCh() chan []byte {
	return p.submitCh
}

// CommitBlock implements AppProxy interface method, calls handler
func (p *InmemAppProxy) CommitBlock(block poset.Block) ([]byte, error) {
	stateHash, err := p.handler.CommitHandler(block)
	p.logger.WithFields(logrus.Fields{
		"round_received": block.RoundReceived(),
		"txs":            len(block.Transactions()),
		"state_hash":     stateHash,
		"err":            err,
	}).Debug("InmemAppProxy.CommitBlock")
	return stateHash, err
}

// GetSnapshot implements AppProxy interface method, calls handler
func (p *InmemAppProxy) GetSnapshot(blockIndex int) ([]byte, error) {
	snapshot, err := p.handler.SnapshotHandler(blockIndex)
	p.logger.WithFields(logrus.Fields{
		"block":    blockIndex,
		"snapshot": snapshot,
		"err":      err,
	}).Debug("InmemAppProxy.GetSnapshot")
	return snapshot, err
}

// Restore implements AppProxy interface method, calls handler
func (p *InmemAppProxy) Restore(snapshot []byte) error {
	stateHash, err := p.handler.RestoreHandler(snapshot)
	p.logger.WithFields(logrus.Fields{
		"state_hash": stateHash,
		"err":        err,
	}).Debug("InmemAppProxy.Restore")
	return err
}

/*
 * staff:
 */

// SubmitTx is called by the App to submit a transaction to Lachesis
func (p *InmemAppProxy) SubmitTx(tx []byte) {
	//have to make a copy, or the tx will be garbage collected and weird stuff
	//happens in transaction pool
	t := make([]byte, len(tx), len(tx))
	copy(t, tx)
	p.submitCh <- t
}
