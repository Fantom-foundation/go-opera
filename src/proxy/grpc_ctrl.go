package proxy

import (
	"context"
	"net"

	"github.com/ethereum/go-ethereum/common"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/inter/pos"
	"github.com/Fantom-foundation/go-lachesis/src/logger"
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
func NewGrpcCtrlProxy(
	bind string, n Node, c Consensus, logger *logrus.Logger, listen network.ListenFunc,
) (
	res CtrlProxy, addr string, err error,
) {
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

// StakeOf returns stake balance of peer.
func (p *grpcCtrlProxy) StakeOf(_ context.Context, req *internal.ID) (*internal.Balance, error) {
	id := common.HexToAddress(req.Hex)
	b := internal.Balance{
		Amount: uint64(p.consensus.StakeOf(id)),
	}

	return &b, nil
}

// GetTxnInfo returns info about transaction.
func (p *grpcCtrlProxy) GetTxnInfo(_ context.Context, req *internal.TransactionRequest) (*internal.TransactionResponse, error) {
	h := hash.HexToTransactionHash(req.Hex)

	var (
		txn   *inter.InternalTransaction
		event *inter.Event
		block *inter.Block
	)

	txn, event = p.node.GetInternalTxn(h)
	if txn == nil {
		return nil, status.Error(codes.NotFound, "transaction not found")
	}

	if event != nil {
		block = p.consensus.GetEventBlock(event.Hash())
	}

	e := event.ToWire()
	return &internal.TransactionResponse{
		Txn:   txn.ToWire(),
		Event: e,
		Block: block.ToWire(),
	}, nil
}

// SendTo makes stake transfer transaction.
// TODO: replace TransferRequest with inter/wire.InternalTransaction
func (p *grpcCtrlProxy) SendTo(_ context.Context, req *internal.TransferRequest) (*internal.TransferResponse, error) {
	tx := inter.InternalTransaction{
		Nonce:      idx.Txn(req.Nonce),
		Amount:     pos.Stake(req.Amount),
		Receiver:   common.HexToAddress(req.Receiver.Hex),
		UntilBlock: idx.Block(req.Until),
	}

	h, err := p.node.AddInternalTxn(tx)

	return &internal.TransferResponse{
		Hex: h.Hex(),
	}, err
}

// SetLogLevel sets logger log level.
func (p *grpcCtrlProxy) SetLogLevel(_ context.Context, req *internal.LogLevel) (*empty.Empty, error) {
	logger.SetLevel(req.Level)
	return &empty.Empty{}, nil
}
