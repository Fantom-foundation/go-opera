package mobile

import (
	"time"

	"github.com/andrecronje/lachesis/src/poset"
	"github.com/mosaicnetworks/babble/src/proxy/inapp"
	"github.com/sirupsen/logrus"
)

/*
This type is not exported
*/

// mobileAppProxy object
type mobileAppProxy struct {
	*inapp.InmemFullProxy

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
	return &mobileAppProxy{
		InmemFullProxy:   inapp.NewInmemFullProxy(time.Second, logger),
		commitHandler:    commitHandler,
		exceptionHandler: exceptionHandler,
		logger:           logger,
	}
}

// CommitBlock commits a Block's to the App and expects the resulting state hash
// gomobile cannot export a Block object because it doesn't support arrays of
// arrays of bytes; so we have to serialize the block.
// Overrides  AppProxy::CommitBlock
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
