package mobile

import (
	"time"

	"github.com/andrecronje/lachesis/src/poset"
	"github.com/mosaicnetworks/babble/src/proxy/inmem"
	"github.com/sirupsen/logrus"
)

/*
This type is not exported
*/

// mobileAppProxy object
type mobileAppProxy struct {
	*inmem.InmemProxy

	commitHandler    CommitHandler
	exceptionHandler ExceptionHandler
	logger           *logrus.Logger
}

// newMobileAppProxy create proxy
func newMobileAppProxy(
	commitHandler CommitHandler,
	exceptionHandler ExceptionHandler,
	logger *logrus.Logger,
) *mobileAppProxy {
	// gomobile cannot export a Block object because it doesn't support arrays of
	// arrays of bytes; so we have to serialize the block.
	commitHandlerFunc := func(block poset.Block) ([]byte, error) {
		blockBytes, err := block.Marshal()
		if err != nil {
			logger.Debug("mobileAppProxy error marhsalling Block")
			return nil, err
		}
		stateHash := commitHandler.OnCommit(blockBytes)
		return stateHash, nil
	}

	snapshotHandlerFunc := func(blockIndex int) ([]byte, error) {
		return []byte{}, nil
	}

	restoreHandlerFunc := func(snapshot []byte) ([]byte, error) {
		return []byte{}, nil
	}

	return &mobileAppProxy{
		InmemProxy: inmem.NewInmemProxy(commitHandlerFunc,
			snapshotHandlerFunc,
			restoreHandlerFunc,
			logger),
		commitHandler:    commitHandler,
		exceptionHandler: exceptionHandler,
		logger:           logger,
	}
}

// CommitBlock commits a Block to the App and expects the resulting state hash
// gomobile cannot export a Block object because it doesn't support arrays of
// arrays of bytes; so we have to serialize the block.
// Overrides  InappProxy::CommitBlock
func (p *mobileAppProxy) CommitBlock(block poset.Block) ([]byte, error) {
	blockBytes, err := block.Marshal()
	if err != nil {
		p.logger.Debug("mobileAppProxy error marhsalling Block")
		return nil, err
	}
	stateHash := p.commitHandler.OnCommit(blockBytes)
	return stateHash, nil
}

//TODO - Implement these two functions
func (p *mobileAppProxy) GetSnapshot(blockIndex int) ([]byte, error) {
	return []byte{}, nil
}

func (p *mobileAppProxy) Restore(snapshot []byte) error {
	return nil
}
