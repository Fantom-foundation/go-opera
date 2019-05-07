package proxy

import (
	"context"
	"net"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/network"
	"github.com/Fantom-foundation/go-lachesis/src/proxy/internal"
)

var (
	// ErrConnTimeout returns when deadline exceeded
	// for Ctrl server connection.
	ErrConnTimeout = errors.New("node connection timeout")
)

// Node representation.
type Node interface {
	GetID() hash.Peer
	AddInternalTxn(inter.InternalTransaction)
}

// Consensus representation.
type Consensus interface {
	GetStakeOf(peer hash.Peer) float64
}

// NewGrpcCtrlProxy starts Ctrl proxy.
func NewGrpcCtrlProxy(bindAddr string, n Node, c Consensus, logger *logrus.Logger, listen network.ListenFunc) (CtrlProxy, error) {
	if logger == nil {
		logger = logrus.New()
		logger.Level = logrus.DebugLevel
	}

	if listen == nil {
		listen = network.TCPListener
	}
	listener := listen(bindAddr)

	s := grpc.NewServer()
	p := GrpcCtrlProxy{
		node:      n,
		consensus: c,
		server:    s,
		listener:  listener,
	}
	internal.RegisterCtrlServer(p.server, &p)

	go func() {
		if err := p.server.Serve(p.listener); err != nil {
			logger.Fatal(err)
		}
	}()

	return &p, nil
}

// GrpcCtrlProxy handles managing requests.
type GrpcCtrlProxy struct {
	node      Node
	consensus Consensus

	server   *grpc.Server
	listener net.Listener
}

// InternalTxn pushes internal transaction into the Node.
func (p *GrpcCtrlProxy) InternalTxn(ctx context.Context, req *internal.InternalTxnRequest) (*empty.Empty, error) {
	peer := hash.HexToPeer(req.Receiver)

	tx := inter.InternalTransaction{
		Amount:   req.Amount,
		Receiver: peer,
	}

	p.node.AddInternalTxn(tx)

	resp := empty.Empty{}
	return &resp, nil
}

// Stake returns the Node stake.
func (p *GrpcCtrlProxy) Stake(ctx context.Context, _ *empty.Empty) (*internal.StakeResponse, error) {
	peer := p.node.GetID()
	resp := internal.StakeResponse{
		Value: p.consensus.GetStakeOf(peer),
	}

	return &resp, nil
}

// ID returns the Node id.
func (p *GrpcCtrlProxy) ID(ctx context.Context, _ *empty.Empty) (*internal.IDResponse, error) {
	peer := p.node.GetID()
	resp := internal.IDResponse{
		Id: peer.Hex(),
	}

	return &resp, nil
}

// Close closes the proxy.
func (p *GrpcCtrlProxy) Close() error {
	p.server.Stop()
	return nil
}
