package proxy

import (
	"context"
	"errors"
	"io"
	"sync/atomic"
	"time"

	"github.com/rs/xid"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/andrecronje/lachesis/src/poset"
	"github.com/andrecronje/lachesis/src/proxy/internal"
	"github.com/andrecronje/lachesis/src/proxy/proto"
)

var (
	ZeroTime        = time.Date(0, time.January, 0, 0, 0, 0, 0, time.Local)
	ErrConnShutdown = errors.New("client disconnected")
)

type GrpcLachesisProxy struct {
	logger    *logrus.Logger
	commitCh  chan proto.Commit
	queryCh   chan proto.SnapshotRequest
	restoreCh chan proto.RestoreRequest

	addr             string
	shutdown         chan struct{}
	reconnect_ticket chan time.Time
	conn             *grpc.ClientConn
	stream           atomic.Value
}

// NewGrpcLachesisProxy instantiates a LachesisProxy-interface connected to remote node
func NewGrpcLachesisProxy(addr string, logger *logrus.Logger) (*GrpcLachesisProxy, error) {
	if logger == nil {
		logger = logrus.New()
		logger.Level = logrus.DebugLevel
	}

	p := &GrpcLachesisProxy{
		addr:             addr,
		shutdown:         make(chan struct{}),
		reconnect_ticket: make(chan time.Time, 1),
		logger:           logger,
		commitCh:         make(chan proto.Commit),
		queryCh:          make(chan proto.SnapshotRequest),
		restoreCh:        make(chan proto.RestoreRequest),
	}

	p.reconnect_ticket <- time.Now()
	err := p.reConnect()
	if err != nil {
		return nil, err
	}

	go p.listen_events()

	return p, nil
}

func (p *GrpcLachesisProxy) Close() error {
	close(p.shutdown)
	close(p.commitCh)
	close(p.queryCh)
	close(p.restoreCh)
	return nil
}

/*
 * inmem interface: LachesisProxy implementation
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
	r := &internal.ToServer{
		Event: &internal.ToServer_Tx_{
			Tx: &internal.ToServer_Tx{
				Data: tx,
			},
		},
	}
	err := p.sendToServer(r)
	return err
}

/*
 * network:
 */

func (p *GrpcLachesisProxy) sendToServer(data *internal.ToServer) (err error) {
	for {
		err = p.getStream().Send(data)
		if err == nil {
			p.logger.Debugf("send to server: %v", data)
			return
		}
		p.logger.Warnf("send to server err: %s", err)

		err = p.reConnect()
		if err == ErrConnShutdown {
			return
		}
	}
}

func (p *GrpcLachesisProxy) recvFromServer() (data *internal.ToClient, err error) {
	for {
		data, err = p.getStream().Recv()
		if err == nil {
			p.logger.Debugf("recv from server: %v", data)
			return
		}
		p.logger.Warnf("recv from server err: %s", err)

		err = p.reConnect()
		if err == ErrConnShutdown {
			return
		}
	}
}

func (p *GrpcLachesisProxy) reConnect() (err error) {
	disconn_time := time.Now()
	connect_time := <-p.reconnect_ticket

	if connect_time == ZeroTime {
		p.reconnect_ticket <- ZeroTime
		return ErrConnShutdown
	}

	if disconn_time.Before(connect_time) {
		p.reconnect_ticket <- connect_time
		return nil
	}

	select {
	case <-p.shutdown:
		p.getStream().CloseSend()
		p.conn.Close()
		p.reconnect_ticket <- ZeroTime
		return ErrConnShutdown
	default:
		// see code below
	}

	p.conn, err = grpc.Dial(p.addr, grpc.WithInsecure())
	if err != nil {
		p.logger.Warnf("dial err: %s", err)
		p.reconnect_ticket <- connect_time
		return
	}

	client := internal.NewLachesisNodeClient(p.conn)
	stream, err := client.Connect(context.TODO())
	if err != nil {
		p.logger.Warnf("rpc Connect() err: %s", err)
		p.reconnect_ticket <- connect_time
		return
	}

	p.setStream(stream)
	p.reconnect_ticket <- time.Now()
	return
}

func (p *GrpcLachesisProxy) listen_events() {
	var (
		event *internal.ToClient
		err   error
		uuid  xid.ID
	)
	for {
		event, err = p.recvFromServer()
		if err != nil {
			if err != io.EOF {
				p.logger.Debugf("recv err: %s", err)
			} else {
				p.logger.Debugf("recv EOF: %s", err)
			}
			break
		}
		// block commit event
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
		// get snapshot query
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
		// restore event
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

/*
 * staff:
 */

func (p *GrpcLachesisProxy) newCommitResponseCh(uuid xid.ID) chan proto.CommitResponse {
	respCh := make(chan proto.CommitResponse)
	go func() {
		var answer *internal.ToServer
		resp, ok := <-respCh
		if ok {
			answer = newAnswer(uuid[:], resp.StateHash, resp.Error)
		}
		p.sendToServer(answer)
	}()
	return respCh
}

func (p *GrpcLachesisProxy) newSnapshotResponseCh(uuid xid.ID) chan proto.SnapshotResponse {
	respCh := make(chan proto.SnapshotResponse)
	go func() {
		var answer *internal.ToServer
		resp, ok := <-respCh
		if ok {
			answer = newAnswer(uuid[:], resp.Snapshot, resp.Error)
		}
		p.sendToServer(answer)
	}()
	return respCh
}

func (p *GrpcLachesisProxy) newRestoreResponseCh(uuid xid.ID) chan proto.RestoreResponse {
	respCh := make(chan proto.RestoreResponse)
	go func() {
		var answer *internal.ToServer
		resp, ok := <-respCh
		if ok {
			answer = newAnswer(uuid[:], resp.StateHash, resp.Error)
		}
		p.sendToServer(answer)
	}()
	return respCh
}

func newAnswer(uuid []byte, data []byte, err error) *internal.ToServer {
	if err != nil {
		return &internal.ToServer{
			Event: &internal.ToServer_Answer_{
				Answer: &internal.ToServer_Answer{
					Uid: uuid,
					Payload: &internal.ToServer_Answer_Error{
						Error: err.Error(),
					},
				},
			},
		}
	}
	return &internal.ToServer{
		Event: &internal.ToServer_Answer_{
			Answer: &internal.ToServer_Answer{
				Uid: uuid,
				Payload: &internal.ToServer_Answer_Data{
					Data: data,
				},
			},
		},
	}
}

func (p *GrpcLachesisProxy) getStream() internal.LachesisNode_ConnectClient {
	return p.stream.Load().(internal.LachesisNode_ConnectClient)
}

func (p *GrpcLachesisProxy) setStream(stream internal.LachesisNode_ConnectClient) {
	p.stream.Store(stream)
}
