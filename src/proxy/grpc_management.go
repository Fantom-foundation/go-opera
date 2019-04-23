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
	"github.com/Fantom-foundation/go-lachesis/src/proxy/proto"
)

const (
	mClientConnTimeout = time.Second
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

// ManagementServer handles managing commands.
type ManagementServer struct {
	Node
	Consensus
}

// InternalTx pushes proto transaction into the node.
func (s *ManagementServer) InternalTx(ctx context.Context, req *proto.InternalTxRequest) (*empty.Empty, error) {
	peer := hash.HexToPeer(req.Receiver)

	tx := inter.InternalTransaction{
		Amount:   req.Amount,
		Receiver: peer,
	}

	s.Node.AddInternalTxn(tx)

	resp := empty.Empty{}
	return &resp, nil
}

// Stake returns node stake.
func (s *ManagementServer) Stake(ctx context.Context, _ *empty.Empty) (*proto.StakeResponse, error) {
	peer := s.Node.GetID()
	resp := proto.StakeResponse{
		Value: float32(s.Consensus.GetStakeOf(peer)),
	}

	return &resp, nil
}

// ID returns node id.
func (s *ManagementServer) ID(ctx context.Context, _ *empty.Empty) (*proto.IDResponse, error) {
	peer := s.Node.GetID()
	resp := proto.IDResponse{
		Id: peer.Hex(),
	}

	return &resp, nil
}

// NewManagmentServer starts managment server.
func NewManagmentServer(bindAddr string, n Node, c Consensus, listen network.ListenFunc) *grpc.Server {
	srv := ManagementServer{
		Node:      n,
		Consensus: c,
	}
	s := grpc.NewServer()
	proto.RegisterManagementServer(s, &srv)

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

// NewManagementClient returns client for lachesis management.
func NewManagementClient(addr string) (proto.ManagementClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mClientConnTimeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, addr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, errors.Wrap(err, "deal context")
	}

	cli := proto.NewManagementClient(conn)
	return cli, nil
}
