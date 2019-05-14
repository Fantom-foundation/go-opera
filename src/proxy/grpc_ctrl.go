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
func (p *grpcCtrlProxy) SelfID(_ context.Context, _ *empty.Empty) (*internal.ID, error) {
	id := p.node.GetID()

	return &internal.ID{
		Hex: id.Hex(),
	}, nil
}

// BalanceOf returns balance of peer.
func (p *grpcCtrlProxy) BalanceOf(_ context.Context, req *internal.ID) (*internal.Balance, error) {
	id := hash.HexToPeer(req.Hex)
	b := internal.Balance{
		Amount: p.consensus.GetBalanceOf(id),
	}

	// Return pending internal transactions if
	// requesting balance of self node.
	if id == p.node.GetID() {
		txs := p.node.GetInternalTxns()

		b.Pending = make([]*internal.Transfer, len(txs))

		for i, tx := range txs {
			b.Pending[i] = &internal.Transfer{
				Amount: tx.Amount,
				Receiver: &internal.ID{
					Hex: tx.Receiver.Hex(),
				},
			}
		}
	}

	return &b, nil
}

// SendTo makes stake transfer transaction.
func (p *grpcCtrlProxy) SendTo(_ context.Context, req *internal.Transfer) (*empty.Empty, error) {
	if req.Amount == 0 {
		return nil, status.Error(codes.InvalidArgument, "cannot transfer zero amount")
	}

	id := p.node.GetID()
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

	p.node.AddInternalTxn(tx)

	return &empty.Empty{}, nil
}
