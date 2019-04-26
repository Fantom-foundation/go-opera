package proxy

import (
	"context"
	"time"

	empty "github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"
	"google.golang.org/grpc"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/network"
	"github.com/Fantom-foundation/go-lachesis/src/proxy/wire"
)

var (
	// ErrConnTimeout returns when deadline exceeded
	// for management server connection.
	ErrConnTimeout = errors.New("node connection timeout")
)

// Node represents node.
type Node interface {
	GetID() hash.Peer
	AddInternalTxn(inter.InternalTransaction)
}

// Consensus represents consensus.
type Consensus interface {
	GetStakeOf(peer hash.Peer) float64
}

// ManagmentServer handles managing requests.
type ManagmentServer struct {
	Node
	Consensus
}

// InternalTx pushes proto transaction into the Node.
func (s *ManagmentServer) InternalTx(ctx context.Context, req *wire.InternalTxRequest) (*empty.Empty, error) {
	peer := hash.HexToPeer(req.Receiver)

	tx := inter.InternalTransaction{
		Amount:   req.Amount,
		Receiver: peer,
	}

	s.Node.AddInternalTxn(tx)

	resp := empty.Empty{}
	return &resp, nil
}

// Stake returns the Node stake.
func (s *ManagmentServer) Stake(ctx context.Context, _ *empty.Empty) (*wire.StakeResponse, error) {
	peer := s.Node.GetID()
	resp := wire.StakeResponse{
		Value: s.Consensus.GetStakeOf(peer),
	}

	return &resp, nil
}

// ID returns the Node id.
func (s *ManagmentServer) ID(ctx context.Context, _ *empty.Empty) (*wire.IDResponse, error) {
	peer := s.Node.GetID()
	resp := wire.IDResponse{
		Id: peer.Hex(),
	}

	return &resp, nil
}

// NewManagmentServer starts managment server.
func NewManagmentServer(bindAddr string, n Node, c Consensus, listen network.ListenFunc) *grpc.Server {
	srv := ManagmentServer{
		Node:      n,
		Consensus: c,
	}
	s := grpc.NewServer()
	wire.RegisterManagmentServer(s, &srv)

	if listen == nil {
		listen = network.TCPListener
	}
	listener := listen(bindAddr)

	go func() {
		if err := s.Serve(listener); err != nil {
			panic(err)
		}
	}()

	return s
}

// NewManagmentClient returns client for lachesis management.
func NewManagmentClient(addr string) (wire.ManagmentClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, addr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		if errors.Cause(err) == context.DeadlineExceeded {
			return nil, ErrConnTimeout
		}
		return nil, err
	}

	cli := wire.NewManagmentClient(conn)
	return cli, nil
}
