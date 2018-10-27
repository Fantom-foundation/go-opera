package proxy

import (
	"context"
	"io"

	"github.com/rs/xid"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/andrecronje/lachesis/src/poset"
	"github.com/andrecronje/lachesis/src/proxy/internal"
	"github.com/andrecronje/lachesis/src/proxy/proto"
)

type GrpcLachesisProxy struct {
	logger    *logrus.Logger
	commitCh  chan proto.Commit
	queryCh   chan proto.SnapshotRequest
	restoreCh chan proto.RestoreRequest

	conn   *grpc.ClientConn
	client internal.LachesisNodeClient
	stream internal.LachesisNode_ConnectClient
}

func NewGrpcLachesisProxy(addr string, logger *logrus.Logger) (*GrpcLachesisProxy, error) {
	var err error

	if logger == nil {
		logger = logrus.New()
		logger.Level = logrus.DebugLevel
	}

	p := &GrpcLachesisProxy{
		logger:    logger,
		commitCh:  make(chan proto.Commit),
		queryCh:   make(chan proto.SnapshotRequest),
		restoreCh: make(chan proto.RestoreRequest),
	}

	// TODO: implement reconnect here
	p.conn, err = grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	p.client = internal.NewLachesisNodeClient(p.conn)

	p.stream, err = p.client.Connect(context.TODO())
	if err != nil {
		return nil, err
	}

	go p.listen_events()

	return p, nil
}

func (p *GrpcLachesisProxy) Close() error {
	_ = p.stream.CloseSend()
	// TODO: log error
	err := p.conn.Close()
	close(p.commitCh)
	close(p.queryCh)
	close(p.restoreCh)
	return err
}

/*
 * inmem interface:
 */

// CommitCh implements LachesisProxy interface method
func (p *GrpcLachesisProxy) CommitCh() chan proto.Commit {
	return p.commitCh
}

// SnapshotRequestCh implements LachesisProxy interface method
func (p *GrpcLachesisProxy) SnapshotRequestCh() chan proto.SnapshotRequest {
	return p.queryCh
}

// RestoreCh implements LachesisProxy interface method
func (p *GrpcLachesisProxy) RestoreCh() chan proto.RestoreRequest {
	return p.restoreCh
}

// SubmitTx implements LachesisProxy interface method
func (p *GrpcLachesisProxy) SubmitTx(tx []byte) error {
	r := &internal.IN{
		Event: &internal.IN_Tx_{
			Tx: &internal.IN_Tx{
				Data: tx,
			},
		},
	}
	err := p.send2server(r)
	return err
}

/*
 * staff:
 */

func (p *GrpcLachesisProxy) listen_events() {
	var (
		event *internal.OUT
		err   error
		uuid  xid.ID
	)
	for {
		event, err = p.stream.Recv()
		// p.logger.Debugf("got event from server: %+v (%s)", event, err)
		if err != nil {
			if err != io.EOF {
				// TODO: log
			}
			// TODO: reconnect
			break
		}

		if b := event.GetBlock(); b != nil {
			var pb poset.Block
			err = pb.Unmarshal(b.Data)
			if err != nil {
				continue
			}
			uuid, err = xid.FromBytes(b.Uid)
			if err == nil {
				p.commitCh <- proto.Commit{
					Block:    pb,
					RespChan: p.newCommitResponseCh(uuid),
				}
			}
			continue
		}

		if q := event.GetQuery(); q != nil {
			uuid, err = xid.FromBytes(q.Uid)
			if err == nil {
				p.queryCh <- proto.SnapshotRequest{
					BlockIndex: int(q.Index),
					RespChan:   p.newSnapshotResponseCh(uuid),
				}
			}
			continue
		}

		if r := event.GetRestore(); r != nil {
			uuid, err = xid.FromBytes(r.Uid)
			if err == nil {
				p.restoreCh <- proto.RestoreRequest{
					Snapshot: r.Data,
					RespChan: p.newRestoreResponseCh(uuid),
				}
			}
			continue
		}
	}
}

func (p *GrpcLachesisProxy) newCommitResponseCh(uuid xid.ID) chan proto.CommitResponse {
	respCh := make(chan proto.CommitResponse)
	go func() {
		var answer *internal.IN
		resp, ok := <-respCh
		if ok {
			answer = newAnswer(uuid[:], resp.StateHash, resp.Error)
		}
		p.send2server(answer)
	}()
	return respCh
}

func (p *GrpcLachesisProxy) newSnapshotResponseCh(uuid xid.ID) chan proto.SnapshotResponse {
	respCh := make(chan proto.SnapshotResponse)
	go func() {
		var answer *internal.IN
		resp, ok := <-respCh
		if ok {
			answer = newAnswer(uuid[:], resp.Snapshot, resp.Error)
		}
		p.send2server(answer)
	}()
	return respCh
}

func (p *GrpcLachesisProxy) newRestoreResponseCh(uuid xid.ID) chan proto.RestoreResponse {
	respCh := make(chan proto.RestoreResponse)
	go func() {
		var answer *internal.IN
		resp, ok := <-respCh
		if ok {
			answer = newAnswer(uuid[:], resp.StateHash, resp.Error)
		}
		p.send2server(answer)
	}()
	return respCh
}

func (p *GrpcLachesisProxy) send2server(data *internal.IN) error {
	return p.stream.Send(data)
}

func newAnswer(uuid []byte, data []byte, err error) *internal.IN {
	if err != nil {
		return &internal.IN{
			Event: &internal.IN_Hash_{
				Hash: &internal.IN_Hash{
					Uid: uuid,
					Answer: &internal.IN_Hash_Error{
						Error: err.Error(),
					},
				},
			},
		}
	}
	return &internal.IN{
		Event: &internal.IN_Hash_{
			Hash: &internal.IN_Hash{
				Uid: uuid,
				Answer: &internal.IN_Hash_Data{
					Data: data,
				},
			},
		},
	}
}
