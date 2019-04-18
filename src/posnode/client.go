package posnode

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"

	"github.com/Fantom-foundation/go-lachesis/src/posnode/api"
)

const (
	tickerInterval = 30 * time.Second
)

type (
	// connection is a wrapper for
	// *grpc.ClientConn
	connection struct {
		*grpc.ClientConn
		initiatedAt time.Time
	}

	// client of node service.
	client struct {
		sync.RWMutex
		connections     map[string]*connection
		connectTimeout  time.Duration
		maxLifeDuration time.Duration
		connWatchTicker *time.Ticker
		opts            []grpc.DialOption
	}
)

func (n *Node) initClient() {
	if n.client.connections != nil {
		return
	}

	n.client.connections = make(map[string]*connection)
	n.client.connectTimeout = n.conf.ConnectTimeout
	n.client.maxLifeDuration = n.conf.ConnectionMaxDuration
	n.client.connWatchTicker = time.NewTicker(tickerInterval)
	go n.client.watchConnections()
}

func (c *connection) shouldReconnect(lifeTime time.Duration) bool {
	return c.initiatedAt.Add(lifeTime).Before(time.Now())
}

func (c *client) watchConnections() {
	for range c.connWatchTicker.C {
		c.RLock()
		connections := c.connections
		c.RUnlock()

		for addr, conn := range connections {
			if conn.GetState() != connectivity.Ready {
				c.remove(addr)
			}
		}
	}
}

func (c *client) get(addr string) (*connection, error) {
	c.RLock()
	conn, exists := c.connections[addr]
	c.RUnlock()

	if !exists {
		c.Lock()
		defer c.Unlock()
		conn, err := c.newConn(addr)
		if err != nil {
			return nil, errors.Wrap(err, "new connection")
		}

		return conn, nil
	}

	if conn.shouldReconnect(c.maxLifeDuration) {
		c.remove(addr)
		return c.get(addr)
	}

	if conn.GetState() != connectivity.Ready {
		c.remove(addr)
		return c.get(addr)
	}

	return conn, nil
}

func (c *client) remove(addr string) {
	c.Lock()
	defer c.Unlock()
	conn := c.connections[addr]
	conn.Close()
	delete(c.connections, addr)
}

func (c *client) newConn(addr string) (*connection, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.connectTimeout)
	defer cancel()

	// TODO: secure connection
	conn, err := grpc.DialContext(ctx, addr, append(c.opts, grpc.WithInsecure(), grpc.WithBlock())...)
	if err != nil {
		return nil, err
	}

	wraper := connection{
		ClientConn:  conn,
		initiatedAt: time.Now(),
	}

	c.connections[addr] = &wraper

	return &wraper, nil
}

// ConnectTo connects to other node service.
func (n *Node) ConnectTo(peer *Peer) (api.NodeClient, error) {
	addr := n.NetAddrOf(peer.Host)
	n.log.Debugf("connect to %s", addr)

	conn, err := n.client.get(addr)
	if err != nil {
		n.log.Warn(errors.Wrapf(err, "connect to: %s", addr))
		return nil, err
	}

	return api.NewNodeClient(conn.ClientConn), nil
}
