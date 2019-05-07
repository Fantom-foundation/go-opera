package proxy

import (
	"context"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"
	"google.golang.org/grpc"

	"github.com/Fantom-foundation/go-lachesis/src/proxy/wire"
)

const (
	commandTimeout = 3 * time.Second
)

// GrpcCmdProxy sends commands to remote node.
type GrpcCmdProxy struct {
	client wire.CtrlClient
}

// GetID returns id of the node.
func (p *GrpcCmdProxy) GetID() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()

	req := empty.Empty{}
	resp, err := p.client.ID(ctx, &req)
	if err != nil {
		return "", err
	}

	return resp.Id, nil
}

// GetStake returns id of the node.
func (p *GrpcCmdProxy) GetStake() (float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()

	req := empty.Empty{}
	resp, err := p.client.Stake(ctx, &req)
	if err != nil {
		return 0, err
	}

	return resp.Value, nil
}

// SubmitInternalTxn adds internal transaction.
func (p *GrpcCmdProxy) SubmitInternalTxn(amount uint64, receiver string) error {
	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()

	req := wire.InternalTxnRequest{
		Amount:   amount,
		Receiver: receiver,
	}

	if _, err := p.client.InternalTxn(ctx, &req); err != nil {
		return err
	}

	return nil
}

// NewGrpcCmdProxy initiates a CmdProxy-interface connected to remote node..
func NewGrpcCmdProxy(addr string, timeout time.Duration) (CmdProxy, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, addr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		if errors.Cause(err) == context.DeadlineExceeded {
			return nil, ErrConnTimeout
		}
		return nil, err
	}

	client := wire.NewCtrlClient(conn)
	proxy := GrpcCmdProxy{
		client: client,
	}

	return &proxy, nil
}
