package proxy

import (
	"context"
	"fmt"
	"net"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/network"
	"github.com/Fantom-foundation/go-lachesis/src/proxy/internal"
)

// grpcCtrlProxy implements CtrlProxy interface.
type grpcCtrlProxy struct {
	node      Node
	consensus Consensus

	server   *grpc.Server
	listener net.Listener
}

// NewGrpcCtrlProxy starts Ctrl proxy.
func NewGrpcCtrlProxy(bind string, n Node, c Consensus, logger *logrus.Logger, listen network.ListenFunc) (
	res CtrlProxy, addr string, err error) {

	if logger == nil {
		logger = logrus.New()
		logger.Level = logrus.DebugLevel
	}

	if listen == nil {
		listen = network.TCPListener
	}
	listener := listen(bind)

	p := &grpcCtrlProxy{
		node:      n,
		consensus: c,
		server:    grpc.NewServer(),
		listener:  listener,
	}
	internal.RegisterNodeServer(p.server, p)

	go func() {
		if err := p.server.Serve(p.listener); err != nil {
			logger.Fatal(err)
		}
	}()

	return p, listener.Addr().String(), nil
}

/*
 * CtrlProxy implementation:
 */

// Close closes the proxy.
func (p *grpcCtrlProxy) Close() {
	p.server.Stop()
}

//TODO: Set descr.
func (p *grpcCtrlProxy) Set() {

}

/*
 * internal.NodeServer implementation:
 */

// ID returns node id.
func (p *grpcCtrlProxy) SelfID(_ context.Context, _ *empty.Empty) (*internal.NodeID, error) {
	id := p.node.GetID()

	return &internal.NodeID{
		Hex: id.Hex(),
	}, nil
}

// BalanceOf returns balance of peer.
func (p *grpcCtrlProxy) BalanceOf(_ context.Context, req *internal.NodeID) (*internal.Balance, error) {
	id := hash.HexToPeer(req.Hex)
	b := internal.Balance{
		Amount: p.consensus.GetBalanceOf(id),
	}

	return &b, nil
}

// Transaction returns infor about transaction.
func (p *grpcCtrlProxy) Transaction(_ context.Context, req *internal.TransactionRequest) (*internal.TransactionResponse, error) {
	h := hash.HexToTransactionHash(req.Hex)
	tx := p.consensus.GetTransaction(h)

	if tx == nil {
		return nil, status.Error(codes.NotFound, "transaction not found")
	}

	return &internal.TransactionResponse{
		Amount: tx.Amount,
		Receiver: &internal.NodeID{
			Hex: tx.Receiver.Hex(),
		},
		Sender: &internal.NodeID{
			Hex: tx.Sender.Hex(),
		},
		Confirmed: tx.Confirmed,
	}, nil
}

// SendTo makes stake transfer transaction.
func (p *grpcCtrlProxy) SendTo(_ context.Context, req *internal.TransferRequest) (*internal.TransferResponse, error) {
	if req.Amount == 0 {
		return nil, status.Error(codes.InvalidArgument, "can not transfer zero amount")
	}

	id := p.node.GetID()
	if id.Hex() == req.Receiver.Hex {
		return nil, status.Error(codes.InvalidArgument, "can not transafer to yourself")
	}

	balance := p.consensus.GetBalanceOf(id)
	if balance < req.Amount {
		message := fmt.Sprintf(
			"insufficient funds %d to transfer %d",
			balance,
			req.Amount,
		)
		return nil, status.Error(codes.InvalidArgument, message)
	}

	tx := inter.InternalTransaction{
		Amount:   req.Amount,
		Receiver: hash.HexToPeer(req.Receiver.Hex),
	}

	h := p.node.AddInternalTxn(tx)

	return &internal.TransferResponse{
		Hex: h.Hex(),
	}, nil
}
