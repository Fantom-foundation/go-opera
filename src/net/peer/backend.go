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

	lnet "github.com/Fantom-foundation/go-lachesis/src/net"
)

const (
	lachesis = "Lachesis"
	// TCP is a Transmission Control Protocol.
	TCP = "tcp"
)

// SyncServer is an interface representing methods for sync server.
type SyncServer interface {
	ReceiverChannel() <-chan *lnet.RPC
	ListenAndServe(network, address string) error
	Close() error
}

// Backend is sync server.
type Backend struct {
	done        chan struct{}
	idleTimeout time.Duration
	listener    *net.TCPListener
	logger      logrus.FieldLogger
	receiver    chan *lnet.RPC
	server      *rpc.Server

	mtx      sync.RWMutex
	shutdown bool

	connsLock sync.Mutex
	conns     map[net.Conn]bool

	wg *sync.WaitGroup
}

// NewBackend creates new sync Backend.
func NewBackend(receiveTimeout, processTimeout, idleTimeout time.Duration,
	logger logrus.FieldLogger) *Backend {
	conns := make(map[net.Conn]bool)
	receiver := make(chan *lnet.RPC)
	done := make(chan struct{})
	rpcServer := rpc.NewServer()
	rpcServer.RegisterName(lachesis, NewLachesis(
		done, receiver, receiveTimeout, processTimeout))

	return &Backend{
		conns:       conns,
		done:        done,
		idleTimeout: idleTimeout,
		logger:      logger,
		receiver:    receiver,
		server:      rpcServer,
		wg:          &sync.WaitGroup{},
	}
}

// ReceiverChannel returns a receiver channel.
func (srv *Backend) ReceiverChannel() <-chan *lnet.RPC {
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
		tcpAddr, err := net.ResolveTCPAddr(network, address)
		if err != nil {
			errChan <- err
			return
		}

		listener, err := net.ListenTCP(network, tcpAddr)
		if err != nil {
			errChan <- err
			return
		}

		srv.listener = listener

		errChan <- nil

		for {
			conn, err := listener.AcceptTCP()
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
		logger.Warn(err)
	}
}

func (srv *Backend) serveCodec(codec rpc.ServerCodec) error {
	defer codec.Close()

	for {
		if err := srv.server.ServeRequest(codec); err != nil {
			return err
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
	srv.listener.Close()

	// Close current connections.
	srv.connsLock.Lock()
	for k := range srv.conns {
		k.Close()
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
			c.Close()
		}
		return
	}
	if err = c.encode(body); err != nil {
		if c.flush() == nil {
			c.Close()
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
	c.rwc.SetDeadline(time.Now().Add(c.idleTimeout))
	return c.dec.Decode(e)
}

func (c *serverCodec) encode(e interface{}) error {
	c.rwc.SetDeadline(time.Now().Add(c.idleTimeout))
	return c.enc.Encode(e)
}

func (c *serverCodec) flush() error {
	return c.encBuf.Flush()
}
