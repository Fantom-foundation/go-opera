package proxy

import (
	"context"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/proxy/internal"
)

const (
	commandTimeout = 3 * time.Second
)

// grpcNodeProxy implements NodeProxy interface.
type grpcNodeProxy struct {
	conn   *grpc.ClientConn
	client internal.NodeClient
	logger *logrus.Logger
}

// NewGrpcNodeProxy initiates a NodeProxy-interface connected to remote node.
func NewGrpcNodeProxy(addr string, logger *logrus.Logger, opts ...grpc.DialOption) (NodeProxy, error) {
	if logger == nil {
		logger = logrus.New()
		logger.Level = logrus.DebugLevel
	}

	p := &grpcNodeProxy{
		logger: logger,
	}

	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	defer cancel()

	var err error
	p.conn, err = grpc.DialContext(ctx, addr,
		append(opts, grpc.WithInsecure(), grpc.WithBlock())...)
	if err != nil {
		return nil, err
	}

	p.client = internal.NewNodeClient(p.conn)

	return p, nil
}

/*
 * NodeProxy implementation:
 */

func (p *grpcNodeProxy) Close() {
	_ = p.conn.Close()
}

func (p *grpcNodeProxy) GetSelfID() (hash.Peer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()

	resp, err := p.client.SelfID(ctx, &empty.Empty{})
	if err != nil {
		return hash.EmptyPeer, err
	}

	return hash.HexToPeer(resp.Hex), nil
}

func (p *grpcNodeProxy) GetBalanceOf(peer hash.Peer) (uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()

	resp, err := p.client.BalanceOf(ctx, &internal.ID{
		Hex: peer.Hex(),
	})
	if err != nil {
		return 0, err
	}

	return resp.Amount, nil
}

func (p *grpcNodeProxy) SendTo(amount uint64, receiver hash.Peer) error {
	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()

	req := internal.Transfer{
		Amount: amount,
		Receiver: &internal.ID{
			Hex: receiver.Hex(),
		},
	}

	if _, err := p.client.SendTo(ctx, &req); err != nil {
		return err
	}

	return nil
}
