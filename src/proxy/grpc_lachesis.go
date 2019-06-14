package proxy

import (
	"context"
	"errors"
	"io"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/xid"
	"google.golang.org/grpc"

	"github.com/Fantom-foundation/go-lachesis/src/logger"
	"github.com/Fantom-foundation/go-lachesis/src/poset"
	"github.com/Fantom-foundation/go-lachesis/src/proxy/internal"
	"github.com/Fantom-foundation/go-lachesis/src/proxy/proto"
)

var (
	errNeedReconnect = errors.New("try to reconnect")
	errConnShutdown  = errors.New("client disconnected")
)

// grpcLachesisProxy implements LachesisProxy interface.
type grpcLachesisProxy struct {
	addr   string
	conn   *grpc.ClientConn
	client internal.LachesisClient
	stream atomic.Value

	reconnTimeout   time.Duration
	reconnectTicket chan time.Time

	commitCh  chan proto.Commit
	queryCh   chan proto.SnapshotRequest
	restoreCh chan proto.RestoreRequest

	shutdown chan struct{}
	wg       sync.WaitGroup

	logger.Instance
}

// NewGrpcLachesisProxy initiates a LachesisProxy-interface connected to remote lachesis node.
func NewGrpcLachesisProxy(addr string, opts ...grpc.DialOption) (LachesisProxy, error) {
	p := &grpcLachesisProxy{
		addr: addr,

		reconnTimeout:   2 * time.Second,
		reconnectTicket: make(chan time.Time, 1),

		commitCh:  make(chan proto.Commit),
		queryCh:   make(chan proto.SnapshotRequest),
		restoreCh: make(chan proto.RestoreRequest),

		shutdown: make(chan struct{}),

		Instance: logger.MakeInstance(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	defer cancel()

	var err error
	p.conn, err = grpc.DialContext(ctx, p.addr,
		append(opts, grpc.WithInsecure(), grpc.WithBackoffMaxDelay(p.reconnTimeout))...)
	if err != nil {
		return nil, err
	}

	p.client = internal.NewLachesisClient(p.conn)

	p.reconnectTicket <- time.Now()

	p.wg.Add(1)
	go p.listenEvents()

	return p, nil
}

func (p *grpcLachesisProxy) Close() {
	close(p.shutdown)

	p.closeStream()
	err := p.conn.Close()

	close(p.commitCh)
	close(p.queryCh)
	close(p.restoreCh)

	if err != nil {
		p.Error(err)
	}

	p.wg.Wait()
}

/*
 * LachesisProxy implementation:
 */

// CommitCh implements LachesisProxy interface method
func (p *grpcLachesisProxy) CommitCh() chan proto.Commit {
	return p.commitCh
}

// SnapshotRequestCh implements LachesisProxy interface method
func (p *grpcLachesisProxy) SnapshotRequestCh() chan proto.SnapshotRequest {
	return p.queryCh
}

// RestoreCh implements LachesisProxy interface method
func (p *grpcLachesisProxy) RestoreCh() chan proto.RestoreRequest {
	return p.restoreCh
}

// SubmitTx implements LachesisProxy interface method
func (p *grpcLachesisProxy) SubmitTx(tx []byte) error {
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

func (p *grpcLachesisProxy) sendToServer(data *internal.ToServer) (err error) {
	for {
		err = p.streamSend(data)
		if err == nil {
			return
		}
		p.Warnf("send to server err: %s", err)

		err = p.reConnect()
		if err == errConnShutdown {
			return
		}
	}
}

func (p *grpcLachesisProxy) recvFromServer() (data *internal.ToClient, err error) {
	for {
		data, err = p.streamRecv()
		if err == nil {
			return
		}
		p.Warnf("recv from server err: %s", err)

		err = p.reConnect()
		if err == errConnShutdown {
			return
		}
	}
}

func (p *grpcLachesisProxy) reConnect() (err error) {
	disconnTime := time.Now()
	connectTime := <-p.reconnectTicket

	select {
	case <-p.shutdown:
		p.reconnectTicket <- connectTime
		return errConnShutdown
	default:
		// see code below
	}

	if disconnTime.Before(connectTime) {
		p.reconnectTicket <- connectTime
		return nil
	}

	var stream internal.Lachesis_ConnectClient
	stream, err = p.client.Connect(
		context.TODO(),
		grpc.MaxCallRecvMsgSize(math.MaxInt32),
		grpc.MaxCallSendMsgSize(math.MaxInt32))
	if err != nil {
		p.Warnf("rpc Connect() err: %s", err)
		time.Sleep(connectTimeout / 2)
		p.reconnectTicket <- connectTime
		return
	}
	p.setStream(stream)

	p.reconnectTicket <- time.Now()
	return
}

func (p *grpcLachesisProxy) listenEvents() {
	defer p.wg.Done()

	var (
		event *internal.ToClient
		err   error
		uuid  xid.ID
	)
	for {
		select {
		case <-p.shutdown:
			return
		default:
			// see code below
		}

		event, err = p.recvFromServer()
		if err != nil {
			if err != io.EOF {
				p.Debugf("recv err: %s", err)
			} else {
				p.Debugf("recv EOF: %s", err)
			}
			break
		}
		// block commit event
		if b := event.GetBlock(); b != nil {
			var pb poset.Block
			err = pb.ProtoUnmarshal(b.Data)
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
					BlockIndex: q.Index,
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

func (p *grpcLachesisProxy) newCommitResponseCh(uuid xid.ID) chan proto.CommitResponse {
	respCh := make(chan proto.CommitResponse)
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()

		var answer *internal.ToServer
		resp, ok := <-respCh
		if ok {
			answer = newAnswer(uuid[:], resp.StateHash, resp.Error)
		}
		if err := p.sendToServer(answer); err != nil {
			p.Debug(err)
		}
	}()
	return respCh
}

func (p *grpcLachesisProxy) newSnapshotResponseCh(uuid xid.ID) chan proto.SnapshotResponse {
	respCh := make(chan proto.SnapshotResponse)
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()

		var answer *internal.ToServer
		resp, ok := <-respCh
		if ok {
			answer = newAnswer(uuid[:], resp.Snapshot, resp.Error)
		}
		if err := p.sendToServer(answer); err != nil {
			p.Debug(err)
		}
	}()
	return respCh
}

func (p *grpcLachesisProxy) newRestoreResponseCh(uuid xid.ID) chan proto.RestoreResponse {
	respCh := make(chan proto.RestoreResponse)
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()

		var answer *internal.ToServer
		resp, ok := <-respCh
		if ok {
			answer = newAnswer(uuid[:], resp.StateHash, resp.Error)
		}
		if err := p.sendToServer(answer); err != nil {
			p.Debug(err)
		}
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

func (p *grpcLachesisProxy) streamSend(data *internal.ToServer) error {
	v := p.stream.Load()
	if v == nil {
		return errNeedReconnect
	}
	stream, ok := v.(internal.Lachesis_ConnectClient)
	if !ok || stream == nil {
		return errNeedReconnect
	}
	return stream.Send(data)
}

func (p *grpcLachesisProxy) streamRecv() (*internal.ToClient, error) {
	v := p.stream.Load()
	if v == nil {
		return nil, errNeedReconnect
	}
	stream, ok := v.(internal.Lachesis_ConnectClient)
	if !ok || stream == nil {
		return nil, errNeedReconnect
	}
	return stream.Recv()
}

func (p *grpcLachesisProxy) setStream(stream internal.Lachesis_ConnectClient) {
	p.stream.Store(stream)
}

func (p *grpcLachesisProxy) closeStream() {
	v := p.stream.Load()
	if v != nil {
		stream, ok := v.(internal.Lachesis_ConnectClient)
		if ok && stream != nil {
			if err := stream.CloseSend(); err != nil {
				p.Debug(err)
			}
		}
	}
}
