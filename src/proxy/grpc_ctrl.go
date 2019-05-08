package proxy

import (
	"context"
	"net"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

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

	return &internal.Balance{
		Amount: p.consensus.GetBalanceOf(id),
	}, nil
}

// SendTo makes stake transfer transaction.
func (p *grpcCtrlProxy) SendTo(_ context.Context, req *internal.Transfer) (*empty.Empty, error) {
	tx := inter.InternalTransaction{
		Amount:   req.Amount,
		Receiver: hash.HexToPeer(req.Receiver.Hex),
	}

	p.node.AddInternalTxn(tx)

	return &empty.Empty{}, nil
}
