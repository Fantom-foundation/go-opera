package peer

import (
	"bufio"
	"encoding/gob"
	"io"
	"net"
	"net/rpc"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/hashicorp/go-multierror"
)

const (
	lachesis = "Lachesis"
	// TCP is a Transmission Control Protocol.
	TCP = "tcp"
)

// CreateListenerFunc creates a new network listener.
type CreateListenerFunc func(network, address string) (net.Listener, error)

// SyncServer is an interface representing methods for sync server.
type SyncServer interface {
	ReceiverChannel() <-chan *RPC
	ListenAndServe(network, address string) error
	Close() error
}

// BackendConfig is a configuration for a sync server.
type BackendConfig struct {
	ReceiveTimeout time.Duration
	ProcessTimeout time.Duration
	IdleTimeout    time.Duration
}

// Backend is sync server.
type Backend struct {
	done         chan struct{}
	idleTimeout  time.Duration
	listener     net.Listener
	listenerFunc CreateListenerFunc
	logger       logrus.FieldLogger
	receiver     chan *RPC
	server       *rpc.Server

	mtx      sync.RWMutex
	shutdown bool

	connsLock sync.Mutex
	conns     map[net.Conn]bool

	wg *sync.WaitGroup
}

// NewBackendConfig creates a default a sync server config.
func NewBackendConfig() *BackendConfig {
	return &BackendConfig{
		// TODO: We increase the values because currently we have too long time for sync process.
		// Revert the values after refactor sync process (node.go)
		ReceiveTimeout: time.Minute * 60,
		ProcessTimeout: time.Minute * 60,
		IdleTimeout:    time.Minute * 10,
	}
}

// NewBackend creates new sync Backend.
func NewBackend(conf *BackendConfig,
	logger logrus.FieldLogger, listenerFunc CreateListenerFunc) *Backend {
	conns := make(map[net.Conn]bool)
	receiver := make(chan *RPC)
	done := make(chan struct{})
	rpcServer := rpc.NewServer()
	if err := rpcServer.RegisterName(lachesis, NewLachesis(
		done, receiver, conf.ReceiveTimeout, conf.ProcessTimeout)); err != nil {
		logger.Panic(err)
	}

	return &Backend{
		conns:        conns,
		done:         done,
		idleTimeout:  conf.IdleTimeout,
		listenerFunc: listenerFunc,
		logger:       logger,
		receiver:     receiver,
		server:       rpcServer,
		wg:           &sync.WaitGroup{},
	}
}

// ReceiverChannel returns a receiver channel.
func (srv *Backend) ReceiverChannel() <-chan *RPC {
	srv.mtx.RLock()
	defer srv.mtx.RUnlock()
	return srv.receiver
}

// ListenAndServe starts sync server.
func (srv *Backend) ListenAndServe(network, address string) error {
	srv.mtx.RLock()
	shutdown := srv.shutdown
	srv.mtx.RUnlock()

	if shutdown {
		return ErrServerAlreadyRunning
	}

	errChan := make(chan error)

	go func() {
		listener, err := srv.listenerFunc(network, address)
		if err != nil {
			errChan <- err
			return
		}

		srv.listener = listener

		errChan <- nil

		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}

			srv.mtx.RLock()
			shutdown := srv.shutdown
			srv.mtx.RUnlock()

			if shutdown {
				return
			}

			srv.wg.Add(1)
			go func() {
				srv.connsLock.Lock()
				srv.conns[conn] = true
				srv.connsLock.Unlock()

				defer func() {
					srv.connsLock.Lock()
					delete(srv.conns, conn)
					srv.connsLock.Unlock()
					srv.wg.Done()
				}()
				srv.serveConn(conn)
			}()
		}
	}()

	return <-errChan
}

//
func (srv *Backend) serveConn(conn net.Conn) {
	logger := srv.logger.WithFields(logrus.Fields{"method": "serveConn",
		"remoteAddr": conn.RemoteAddr().String()})

	buf := bufio.NewWriter(conn)
	codec := &serverCodec{
		rwc:         conn,
		dec:         gob.NewDecoder(conn),
		enc:         gob.NewEncoder(buf),
		encBuf:      buf,
		idleTimeout: srv.idleTimeout,
	}
	// Set idle timeout.
	if err := codec.rwc.SetDeadline(
		time.Now().Add(srv.idleTimeout)); err != nil {
		logger.Error(err)
		return
	}

	if err := srv.serveCodec(codec); err != io.EOF {
		logger.Warn(err.GoString())
	}
}

func (srv *Backend) serveCodec(codec rpc.ServerCodec) *multierror.Error {
	var result *multierror.Error
	defer func() {
		if err := codec.Close(); err != nil {
			result = multierror.Append(result, err)
			println(err.Error())
		}
	}()

	for {
		if err := srv.server.ServeRequest(codec); err != nil {
			result = multierror.Append(result, err)
			println(err.Error())
			return result
		}
	}
}

// Close stops sync server.
func (srv *Backend) Close() error {
	srv.mtx.Lock()
	defer srv.mtx.Unlock()

	if srv.shutdown {
		return nil
	}

	srv.shutdown = true

	if srv.listener == nil {
		return nil
	}

	// Stop accepting new connections.
	if err := srv.listener.Close(); err != nil {
		return err
	}

	// Close current connections.
	srv.connsLock.Lock()
	var er error
	for k := range srv.conns {
		if err := k.Close(); err != nil {
			er = err
		}
	}
	if er != nil {
		return er
	}
	srv.connsLock.Unlock()

	// Stop handler.
	close(srv.done)

	// Wait for all connections to complete.
	srv.wg.Wait()
	return nil
}

type serverCodec struct {
	rwc         net.Conn
	dec         *gob.Decoder
	enc         *gob.Encoder
	encBuf      *bufio.Writer
	idleTimeout time.Duration
	closed      bool
}

func (c *serverCodec) ReadRequestHeader(r *rpc.Request) error {
	return c.decode(r)
}

func (c *serverCodec) ReadRequestBody(body interface{}) error {
	return c.decode(body)
}

func (c *serverCodec) WriteResponse(
	r *rpc.Response, body interface{}) (err error) {
	if err = c.encode(r); err != nil {
		if c.flush() == nil {
			if err := c.Close(); err != nil {
				return err
			}
		}
		return
	}
	if err = c.encode(body); err != nil {
		if c.flush() == nil {
			if err := c.Close(); err != nil {
				return err
			}
		}
		return
	}
	return c.flush()
}

func (c *serverCodec) Close() error {
	if c.closed {
		return nil
	}
	c.closed = true
	return c.rwc.Close()
}

func (c *serverCodec) decode(e interface{}) error {
	err := c.rwc.SetDeadline(time.Now().Add(c.idleTimeout))
	if err == nil {
		err = c.dec.Decode(e)
	}
	return err
}

func (c *serverCodec) encode(e interface{}) error {
	err := c.rwc.SetDeadline(time.Now().Add(c.idleTimeout))
	if err == nil {
		err = c.enc.Encode(e)
	}
	return err
}

func (c *serverCodec) flush() error {
	return c.encBuf.Flush()
}
