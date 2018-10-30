package dummy

import (
	"github.com/sirupsen/logrus"

	"github.com/andrecronje/lachesis/src/proxy"
)

// DummyClient is a implementation of the dummy app. Lachesis and the
// app run in separate processes and communicate through proxy
type DummyClient struct {
	logger        *logrus.Logger
	state         proxy.ProxyHandler
	lachesisProxy proxy.LachesisProxy
}

func NewInmemDummyApp(logger *logrus.Logger) proxy.AppProxy {
	state := NewState(logger)
	return proxy.NewInmemAppProxy(state, logger)
}

func NewDummySocketClient(addr string, logger *logrus.Logger) (*DummyClient, error) {
	lachesisProxy, err := proxy.NewGrpcLachesisProxy(addr, logger)
	if err != nil {
		return nil, err
	}

	return NewDummyClient(lachesisProxy, nil, logger)
}

// NewDummyClient instantiates an implementation of the dummy app
func NewDummyClient(lachesisProxy proxy.LachesisProxy, handler proxy.ProxyHandler, logger *logrus.Logger) (c *DummyClient, err error) {
	// state := NewState(logger)

	c = &DummyClient{
		logger:        logger,
		state:         handler,
		lachesisProxy: lachesisProxy,
	}

	if handler == nil {
		return
	}

	go func() {
		for {
			select {

			case b, ok := <-lachesisProxy.CommitCh():
				if !ok {
					return
				}
				logger.Debugf("block commit event: %v", b.Block)
				hash, err := handler.CommitHandler(b.Block)
				b.Respond(hash, err)

			case r, ok := <-lachesisProxy.RestoreCh():
				if !ok {
					return
				}
				logger.Debugf("snapshot restore command: %v", r.Snapshot)
				hash, err := handler.RestoreHandler(r.Snapshot)
				r.Respond(hash, err)

			case s, ok := <-lachesisProxy.SnapshotRequestCh():
				if !ok {
					return
				}
				logger.Debugf("get snapshot query: %v", s.BlockIndex)
				hash, err := handler.SnapshotHandler(s.BlockIndex)
				s.Respond(hash, err)
			}
		}
	}()

	return
}

// SubmitTx sends a transaction to node via proxy
func (c *DummyClient) SubmitTx(tx []byte) error {
	return c.lachesisProxy.SubmitTx(tx)
}
