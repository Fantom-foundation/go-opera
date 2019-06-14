package dummy

import (
	"github.com/sirupsen/logrus"

	"github.com/Fantom-foundation/go-lachesis/src/proxy"
)

// Client is a implementation of the dummy app. Lachesis and the
// app run in separate processes and communicate through proxy
type Client struct {
	logger        *logrus.Logger
	state         proxy.App
	lachesisProxy proxy.LachesisProxy
}

// NewInmemApp constructor
func NewInmemApp(logger *logrus.Logger) proxy.AppProxy {
	state := NewState(logger)
	return proxy.NewInmemAppProxy(state)
}

// NewSocketClient constructor
func NewSocketClient(addr string, logger *logrus.Logger) (*Client, error) {
	lachesisProxy, err := proxy.NewGrpcLachesisProxy(addr)
	if err != nil {
		return nil, err
	}

	return NewClient(lachesisProxy, nil, logger)
}

// NewClient instantiates an implementation of the dummy app
func NewClient(lachesisProxy proxy.LachesisProxy, handler proxy.App, logger *logrus.Logger) (c *Client, err error) {
	// state := NewState(logger)

	c = &Client{
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
func (c *Client) SubmitTx(tx []byte) error {
	return c.lachesisProxy.SubmitTx(tx)
}
