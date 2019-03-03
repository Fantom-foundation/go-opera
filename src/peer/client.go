package peer

import (
	"context"
	"net"
	"net/rpc"
	"time"
)

// CreateSyncClientFunc is a function to create a sync client.
type CreateSyncClientFunc func(target string,
	timeout time.Duration) (SyncClient, error)

// CreateNetConnFunc is a function to create new network connection.
type CreateNetConnFunc func(network,
	address string, timeout time.Duration) (net.Conn, error)

// RPCClient is an interface representing methods for a RPC Client.
type RPCClient interface {
	Go(serviceMethod string, args interface{},
		reply interface{}, done chan *rpc.Call) *rpc.Call
	Close() error
}

// SyncClient is an interface representing methods for sync client.
type SyncClient interface {
	Sync(ctx context.Context,
		req *SyncRequest, resp *SyncResponse) error
	ForceSync(ctx context.Context,
		req *ForceSyncRequest, resp *ForceSyncResponse) error
	FastForward(ctx context.Context,
		req *FastForwardRequest, resp *FastForwardResponse) error
	Close() error
}

// Client is a sync client.
type Client struct {
	connect RPCClient
}

// NewRPCClient creates new RPC client.
func NewRPCClient(
	network, address string, timeout time.Duration,
	createNetConnFunc CreateNetConnFunc) (*rpc.Client, error) {
	conn, err := createNetConnFunc(network, address, timeout)
	if err != nil {
		return nil, err
	}

	return rpc.NewClient(conn), nil
}

// NewClient creates new sync client.
func NewClient(rpcClient RPCClient) (*Client, error) {
	return &Client{connect: rpcClient}, nil
}

// Sync sends a sync request.
func (c *Client) Sync(ctx context.Context,
	req *SyncRequest, resp *SyncResponse) error {
	return c.call(ctx, MethodSync, req, resp, nil)
}

// ForceSync sends a force sync request.
func (c *Client) ForceSync(ctx context.Context,
	req *ForceSyncRequest, resp *ForceSyncResponse) error {
	return c.call(ctx, MethodForceSync, req, resp, nil)
}

// FastForward sends a fast forward request.
func (c *Client) FastForward(ctx context.Context,
	req *FastForwardRequest, resp *FastForwardResponse) error {
	return c.call(ctx, MethodFastForward, req, resp, nil)
}

// Close closes a sync client.
func (c *Client) Close() error {
	return c.connect.Close()
}

func (c *Client) call(ctx context.Context, serviceMethod string,
	req interface{}, resp interface{}, done chan *rpc.Call) error {
	call := c.connect.Go(serviceMethod, req, resp, nil)

	select {
	case replay := <-call.Done:
		if replay.Error != nil {
			return replay.Error
		}
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}
