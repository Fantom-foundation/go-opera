package proxy

//go:generate protoc --go_out=plugins=grpc:./ ./internal/grpc.proto
// Install before go generate:
//  wget https://github.com/protocolbuffers/protobuf/releases/download/v3.6.1/protoc-3.6.1-linux-x86_64.zip
//  unzip protoc-3.6.1-linux-x86_64.zip -x readme.txt -d /usr/local/
//  go get -u github.com/golang/protobuf/protoc-gen-go

import (
	"errors"
	"io"
	"net"
	"sync"
	"time"

	"github.com/rs/xid"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/andrecronje/lachesis/src/poset"
	"github.com/andrecronje/lachesis/src/proxy/internal"
)

var ErrNoAnswers = errors.New("No answers")

type ClientStream internal.LachesisNode_ConnectServer

//GrpcAppProxy implements the AppProxy interface
type GrpcAppProxy struct {
	logger   *logrus.Logger
	listener net.Listener
	server   *grpc.Server

	timeout      time.Duration
	new_clients  chan ClientStream
	askings      map[xid.ID]chan *internal.ToServer_Answer
	askings_sync sync.RWMutex

	event4server  chan []byte
	event4clients chan *internal.ToClient
}

// NewGrpcAppProxy instantiates a joined AppProxy-interface listen to remote apps
func NewGrpcAppProxy(bind_addr string, timeout time.Duration, logger *logrus.Logger) (*GrpcAppProxy, error) {
	var err error

	if logger == nil {
		logger = logrus.New()
		logger.Level = logrus.DebugLevel
	}

	// TODO: make it buffered?
	p := &GrpcAppProxy{
		logger:        logger,
		timeout:       timeout,
		new_clients:   make(chan ClientStream, 100),
		askings:       make(map[xid.ID]chan *internal.ToServer_Answer),
		event4server:  make(chan []byte),
		event4clients: make(chan *internal.ToClient),
	}

	p.listener, err = net.Listen("tcp", bind_addr)
	if err != nil {
		return nil, err
	}
	p.server = grpc.NewServer()
	internal.RegisterLachesisNodeServer(p.server, p)
	go p.server.Serve(p.listener)

	go p.send_events4clients()

	return p, nil
}

func (p *GrpcAppProxy) Close() error {
	p.server.Stop()
	p.listener.Close()
	close(p.event4server)
	close(p.event4clients)
	return nil
}

/*
 * network interface:
 */

// Connect implements gRPC-server interface: LachesisNodeServer
func (p *GrpcAppProxy) Connect(stream internal.LachesisNode_ConnectServer) error {
	// save client's stream for writing
	p.new_clients <- stream
	p.logger.Debugf("client connected")
	// read from stream
	for {
		req, err := stream.Recv()
		if err != nil {
			if err != io.EOF {
				p.logger.Debugf("client refused: %s", err)
			} else {
				p.logger.Debugf("client disconnected")
			}
			return err
		}
		if tx := req.GetTx(); tx != nil {
			p.event4server <- tx.GetData()
			continue
		}
		if answer := req.GetAnswer(); answer != nil {
			p.route_answer(answer)
			continue
		}
	}

	return nil
}

func (p *GrpcAppProxy) send_events4clients() {
	var (
		err       error
		connected []ClientStream
		alive     []ClientStream
		stream    ClientStream
	)
	for event := range p.event4clients {

		for i := len(p.new_clients); i > 0; i-- {
			stream = <-p.new_clients
			connected = append(connected, stream)
		}

		for _, stream = range connected {
			err = stream.Send(event)
			if err == nil {
				alive = append(alive, stream)
			}
		}

		connected = alive
		alive = nil
	}
}

/*
 * inmem interface: AppProxy implementation
 */

// SubmitCh implements AppProxy interface method
func (p *GrpcAppProxy) SubmitCh() chan []byte {
	return p.event4server
}

// CommitBlock implements AppProxy interface method
func (p *GrpcAppProxy) CommitBlock(block poset.Block) ([]byte, error) {
	data, err := block.Marshal()
	if err != nil {
		return nil, err
	}
	answer, ok := <-p.push_block(data)
	if !ok {
		return nil, ErrNoAnswers
	}
	err_msg := answer.GetError()
	if err_msg != "" {
		return nil, errors.New(err_msg)
	}
	return answer.GetData(), nil
}

// GetSnapshot implements AppProxy interface method
func (p *GrpcAppProxy) GetSnapshot(blockIndex int) ([]byte, error) {
	answer, ok := <-p.push_query(int64(blockIndex))
	if !ok {
		return nil, ErrNoAnswers
	}
	err_msg := answer.GetError()
	if err_msg != "" {
		return nil, errors.New(err_msg)
	}
	return answer.GetData(), nil
}

// Restore implements AppProxy interface method
func (p *GrpcAppProxy) Restore(snapshot []byte) error {
	answer, ok := <-p.push_restore(snapshot)
	if !ok {
		return ErrNoAnswers
	}
	err_msg := answer.GetError()
	if err_msg != "" {
		return errors.New(err_msg)
	}
	return nil
}

/*
 * staff:
 */

func (p *GrpcAppProxy) route_answer(hash *internal.ToServer_Answer) {
	uuid, err := xid.FromBytes(hash.GetUid())
	if err != nil {
		// TODO: log invalid uuid
		return
	}
	p.askings_sync.RLock()
	if ch, ok := p.askings[uuid]; ok {
		ch <- hash
	}
	p.askings_sync.RUnlock()
}

func (p *GrpcAppProxy) push_block(block []byte) chan *internal.ToServer_Answer {
	uuid := xid.New()
	event := &internal.ToClient{
		Event: &internal.ToClient_Block_{
			Block: &internal.ToClient_Block{
				Uid:  uuid[:],
				Data: block,
			},
		},
	}
	answer := p.subscribe4answer(uuid)
	p.event4clients <- event
	return answer
}

func (p *GrpcAppProxy) push_query(index int64) chan *internal.ToServer_Answer {
	uuid := xid.New()
	event := &internal.ToClient{
		Event: &internal.ToClient_Query_{
			Query: &internal.ToClient_Query{
				Uid:   uuid[:],
				Index: index,
			},
		},
	}
	answer := p.subscribe4answer(uuid)
	p.event4clients <- event
	return answer
}

func (p *GrpcAppProxy) push_restore(snapshot []byte) chan *internal.ToServer_Answer {
	uuid := xid.New()
	event := &internal.ToClient{
		Event: &internal.ToClient_Restore_{
			Restore: &internal.ToClient_Restore{
				Uid:  uuid[:],
				Data: snapshot,
			},
		},
	}
	answer := p.subscribe4answer(uuid)
	p.event4clients <- event
	return answer
}

func (p *GrpcAppProxy) subscribe4answer(uuid xid.ID) chan *internal.ToServer_Answer {
	ch := make(chan *internal.ToServer_Answer)
	p.askings_sync.Lock()
	p.askings[uuid] = ch
	p.askings_sync.Unlock()
	// timeout
	go func() {
		<-time.After(p.timeout)
		p.askings_sync.Lock()
		delete(p.askings, uuid)
		p.askings_sync.Unlock()
		close(ch)
	}()

	return ch
}
