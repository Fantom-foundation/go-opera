package proxy

import (
	"context"
	"errors"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/inter/pos"
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

func (p *grpcNodeProxy) GetSelfID() (common.Address, error) {
	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()

	resp, err := p.client.SelfID(ctx, &empty.Empty{})
	if err != nil {
		return common.Address{}, unwrapGrpcErr(err)
	}

	return common.HexToAddress(resp.Hex), nil
}

func (p *grpcNodeProxy) StakeOf(peer common.Address) (pos.Stake, error) {
	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()

	resp, err := p.client.StakeOf(ctx, &internal.ID{
		Hex: peer.Hex(),
	})
	if err != nil {
		return pos.Stake(0), unwrapGrpcErr(err)
	}

	return pos.Stake(resp.Amount), nil
}

func (p *grpcNodeProxy) SendTo(receiver common.Address, index idx.Txn, amount pos.Stake, until idx.Block) (hash.Transaction, error) {
	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()

	req := internal.TransferRequest{
		Receiver: &internal.ID{
			Hex: receiver.Hex(),
		},
		Nonce:  uint64(index),
		Amount: uint64(amount),
		Until:  uint64(until),
	}

	resp, err := p.client.SendTo(ctx, &req)
	if err != nil {
		return hash.ZeroTransaction, unwrapGrpcErr(err)
	}

	return hash.HexToTransactionHash(resp.Hex), nil
}

func (p *grpcNodeProxy) GetTxnInfo(t hash.Transaction) (*inter.InternalTransaction, *inter.Event, *inter.Block, error) {
	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()

	req := internal.TransactionRequest{
		Hex: t.Hex(),
	}

	resp, err := p.client.GetTxnInfo(ctx, &req)
	if err != nil {
		return nil, nil, nil, unwrapGrpcErr(err)
	}

	return inter.WireToInternalTransaction(resp.Txn),
		inter.WireToEvent(resp.Event),
		inter.WireToBlock(resp.Block),
		nil
}

func (p *grpcNodeProxy) SetLogLevel(l string) error {
	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()

	req := internal.LogLevel{
		Level: l,
	}

	if _, err := p.client.SetLogLevel(ctx, &req); err != nil {
		return unwrapGrpcErr(err)
	}

	return nil
}

func unwrapGrpcErr(err error) error {
	st := status.Convert(err)
	return errors.New(st.Message())

}
