package mobile

import (
	"github.com/andrecronje/lachesis/src/poset"
	"github.com/andrecronje/lachesis/src/proxy/inmem"
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

	mobileApp := &mobileAppProxy{
		commitHandler:    commitHandler,
		exceptionHandler: exceptionHandler,
		logger:           logger,
	}

	mobileApp.InmemProxy = inmem.NewInmemProxy(mobileApp, logger)

	return mobileApp
}

func (m *mobileAppProxy) CommitHandler(block poset.Block) ([]byte, error) {
	blockBytes, err := block.Marshal()
 	if err != nil {
		m.logger.Debug("mobileAppProxy error marhsalling Block")
 		return nil, err
	}
 	stateHash := m.commitHandler.OnCommit(blockBytes)
 	return stateHash, nil
}
 func (m *mobileAppProxy) SnapshotHandler(blockIndex int) ([]byte, error) {
	return []byte{}, nil
}
 func (m *mobileAppProxy) RestoreHandler(snapshot []byte) ([]byte, error) {
	return []byte{}, nil
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
