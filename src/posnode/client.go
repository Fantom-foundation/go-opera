package posnode

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/grpc"

	"github.com/Fantom-foundation/go-lachesis/src/posnode/api"
)

type (
	// connections pool for client.
	connPool struct {
		sync.RWMutex
		connections map[string]*grpc.ClientConn
		timeout     time.Duration
		opts        []grpc.DialOption
	}

	// client of node service.
	client struct {
		pool *connPool
	}
)

func newClient(timeout time.Duration, opts []grpc.DialOption) client {
	pool := connPool{
		RWMutex:     sync.RWMutex{},
		connections: make(map[string]*grpc.ClientConn),
		timeout:     timeout,
		opts:        opts,
	}

	cli := client{
		pool: &pool,
	}

	return cli
}

func (p *connPool) get(addr string) (*grpc.ClientConn, error) {
	p.RLock()
	conn, exists := p.connections[addr]
	p.RUnlock()
	if !exists {
		p.Lock()
		defer p.Unlock()
		conn, err := p.newConn(addr, p.opts...)
		if err != nil {
			return nil, errors.Wrap(err, "new conn")
		}
		p.connections[addr] = conn

		return conn, nil
	}

	return conn, nil
}

func (p *connPool) newConn(addr string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), p.timeout)
	defer cancel()

	// TODO: secure connection
	conn, err := grpc.DialContext(ctx, addr, append(opts, grpc.WithInsecure(), grpc.WithBlock())...)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// ConnectTo connects to other node service.
func (n *Node) ConnectTo(peer *Peer) (api.NodeClient, error) {
	addr := n.NetAddrOf(peer.Host)
	n.log.Debugf("connect to %s", addr)

	conn, err := n.pool.get(addr)
	if err != nil {
		n.log.Warn(errors.Wrapf(err, "connect to: %s", addr))
		return nil, err
	}

	return api.NewNodeClient(conn), nil
}
