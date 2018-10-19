package app

import (
	"fmt"

	bcrypto "github.com/andrecronje/lachesis/src/crypto"
	"github.com/andrecronje/lachesis/src/log"
	"github.com/andrecronje/lachesis/src/poset"
	"github.com/sirupsen/logrus"
)

//InmemProxy is used for testing
type InmemAppProxy struct {
	submitCh              chan []byte
	stateHash             []byte
	committedTransactions [][]byte
	snapshots             map[int][]byte
	logger                *logrus.Logger
}

func NewInmemAppProxy(logger *logrus.Logger) *InmemAppProxy {
	if logger == nil {
		logger = logrus.New()
		logger.Level = logrus.DebugLevel
		lachesis_log.NewLocal(logger, logger.Level.String())
	}

	return &InmemAppProxy{
		submitCh:              make(chan []byte),
		stateHash:             []byte{},
		committedTransactions: [][]byte{},
		snapshots:             make(map[int][]byte),
		logger:                logger,
	}
}

func (p *InmemAppProxy) commit(block poset.Block) ([]byte, error) {
	p.committedTransactions = append(p.committedTransactions, block.Transactions()...)

	hash := p.stateHash

	for _, t := range block.Transactions() {
		tHash := bcrypto.SHA256(t)
		hash = bcrypto.SimpleHashFromTwoHashes(hash, tHash)
	}

	p.stateHash = hash

	//XXX do something smart here
	p.snapshots[block.Index()] = hash

	return p.stateHash, nil
}

func (p *InmemAppProxy) restore(snapshot []byte) error {
	//XXX do something smart here
	p.stateHash = snapshot

	return nil
}

//------------------------------------------------------------------------------
//Implement AppProxy Interface

func (p *InmemAppProxy) SubmitCh() chan []byte {
	return p.submitCh
}

func (p *InmemAppProxy) CommitBlock(block poset.Block) (stateHash []byte, err error) {
	p.logger.WithFields(logrus.Fields{
		"round_received": block.RoundReceived(),
		"txs":            len(block.Transactions()),
	}).Debug("InmemProxy CommitBlock")

	return p.commit(block)
}

func (p *InmemAppProxy) GetSnapshot(blockIndex int) (snapshot []byte, err error) {
	p.logger.WithField("block", blockIndex).Debug("InmemProxy GetSnapshot")

	snapshot, ok := p.snapshots[blockIndex]

	if !ok {
		return nil, fmt.Errorf("Snapshot %d not found", blockIndex)
	}

	return snapshot, nil
}

func (p *InmemAppProxy) Restore(snapshot []byte) error {
	p.logger.WithField("snapshot", snapshot).Debug("Restore")

	return p.restore(snapshot)
}

//------------------------------------------------------------------------------

func (p *InmemAppProxy) SubmitTx(tx []byte) {
	p.submitCh <- tx
}

func (p *InmemAppProxy) GetCommittedTransactions() [][]byte {
	return p.committedTransactions
}
