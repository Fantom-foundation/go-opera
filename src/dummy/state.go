package dummy

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/andrecronje/lachesis/src/crypto"
	"github.com/andrecronje/lachesis/src/poset"
)

/*
 * The dummy App is used for testing and as an example for building Lachesis
 * applications. Here, we define the dummy's state which doesn't really do
 * anything useful. It saves and logs block transactions. The state hash is
 * computed by cumulatively hashing transactions together as they come in.
 * Snapshots correspond to the state hash resulting from executing a the block's
 * transactions.
 */

// State implements ProxyHandler
type State struct {
	logger       *logrus.Logger
	committedTxs [][]byte
	stateHash    []byte
	snapshots    map[int][]byte
}

func NewState(logger *logrus.Logger) *State {
	state := &State{
		logger:       logger,
		committedTxs: [][]byte{},
		stateHash:    []byte{},
		snapshots:    make(map[int][]byte),
	}
	logger.Info("Init Dummy State")

	return state
}

/*
 * inmem interface: ProxyHandler implementation
 */

func (s *State) CommitHandler(block poset.Block) ([]byte, error) {
	s.logger.WithField("block", block).Debug("CommitBlock")

	err := s.commit(block)
	if err != nil {
		return nil, err
	}
	s.logger.WithField("stateHash", s.stateHash).Debug("CommitBlock Answer")
	return s.stateHash, nil
}

func (s *State) SnapshotHandler(blockIndex int) ([]byte, error) {
	s.logger.WithField("block", blockIndex).Debug("GetSnapshot")

	snapshot, ok := s.snapshots[blockIndex]
	if !ok {
		return nil, fmt.Errorf("Snapshot %d not found", blockIndex)
	}

	return snapshot, nil
}

func (s *State) RestoreHandler(snapshot []byte) ([]byte, error) {
	//XXX do something smart here
	s.stateHash = snapshot
	return s.stateHash, nil
}

/*
 * staff:
 */

func (s *State) GetCommittedTransactions() [][]byte {
	return s.committedTxs
}

func (s *State) commit(block poset.Block) error {
	s.committedTxs = append(s.committedTxs, block.Transactions()...)
	// log tx and update state hash
	hash := s.stateHash
	for _, tx := range block.Transactions() {
		s.logger.Info(string(tx))
		hash = crypto.SimpleHashFromTwoHashes(hash, crypto.SHA256(tx))
	}
	s.snapshots[block.Index()] = hash
	s.stateHash = hash
	return nil
}
