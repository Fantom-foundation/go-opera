package proxy

//go:generate protoc --go_out=plugins=grpc:./ ./internal/grpc.proto
// Install before go generate:
//  wget https://github.com/protocolbuffers/protobuf/releases/download/v3.6.1/protoc-3.6.1-linux-x86_64.zip
//  unzip protoc-3.6.1-linux-x86_64.zip -x readme.txt -d /usr/local/
//  go get -u github.com/golang/protobuf/protoc-gen-go

import (
	"errors"
	"io"
	"math"
	"net"
	"sync"
	"time"

	"github.com/rs/xid"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/Fantom-foundation/go-lachesis/src/poset"
	"github.com/Fantom-foundation/go-lachesis/src/proxy/internal"
)

var ErrNoAnswers = errors.New("no answers")

type ClientStream internal.LachesisNode_ConnectServer

//GrpcAppProxy implements the AppProxy interface
type GrpcAppProxy struct {
	logger   *logrus.Logger
	listener net.Listener
	server   *grpc.Server

	timeout     time.Duration
	newClients  chan ClientStream
	askings     map[xid.ID]chan *internal.ToServer_Answer
	askingsSync sync.RWMutex

	event4server  chan []byte
	event4clients chan *internal.ToClient
}

// NewGrpcAppProxy instantiates a joined AppProxy-interface listen to remote apps
func NewGrpcAppProxy(bindAddr string, timeout time.Duration, logger *logrus.Logger) (*GrpcAppProxy, error) {
	var err error

	if logger == nil {
		logger = logrus.New()
		logger.Level = logrus.DebugLevel
	}

	p := &GrpcAppProxy{
		logger:     logger,
		timeout:    timeout,
		newClients: make(chan ClientStream, 100),
		// TODO: make chans buffered?
		askings:       make(map[xid.ID]chan *internal.ToServer_Answer),
		event4server:  make(chan []byte),
		event4clients: make(chan *internal.ToClient),
	}

	p.listener, err = net.Listen("tcp", bindAddr)
	if err != nil {
		return nil, err
	}
	p.server = grpc.NewServer(
		grpc.MaxRecvMsgSize(math.MaxInt32),
		grpc.MaxSendMsgSize(math.MaxInt32))
	internal.RegisterLachesisNodeServer(p.server, p)

	go func() {
		if err := p.server.Serve(p.listener); err != nil {
			logger.Fatal(err)
		}
	}()

	go p.sendEvents4clients()

	return p, nil
}

func (p *GrpcAppProxy) Close() error {
	p.server.Stop()
	//All listeners are closed by gRPC.Stop() function
	//err := p.listener.Close()
	close(p.event4server)
	close(p.event4clients)
	return nil //err
}

/*
 * network interface:
 */

// Connect implements gRPC-server interface: LachesisNodeServer
func (p *GrpcAppProxy) Connect(stream internal.LachesisNode_ConnectServer) error {
	// save client's stream for writing
	p.newClients <- stream
	p.logger.Debugf("client connected")
	// read from stream
	for {
		req, err := stream.Recv()
		if err != nil {
			if err != io.EOF {
				p.logger.Debugf("client refused: %s", err)
			} else {
				p.logger.Debugf("client disconnected well")
			}
			return err
		}
		if tx := req.GetTx(); tx != nil {
			p.event4server <- tx.GetData()
			continue
		}
		if answer := req.GetAnswer(); answer != nil {
			p.routeAnswer(answer)
			continue
		}
	}
}

func (p *GrpcAppProxy) sendEvents4clients() {
	var (
		err       error
		connected []ClientStream
		alive     []ClientStream
		stream    ClientStream
	)
	for event := range p.event4clients {

		for i := len(p.newClients); i > 0; i-- {
			stream = <-p.newClients
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

// SubmitCh implements AppProxy interface method
// TODO: Incorrect implementation, just adding to the interface so long
func (p *GrpcAppProxy) SubmitInternalCh() chan poset.InternalTransaction {
	return nil
}

// CommitBlock implements AppProxy interface method
func (p *GrpcAppProxy) CommitBlock(block poset.Block) ([]byte, error) {
	data, err := block.ProtoMarshal()
	if err != nil {
		return nil, err
	}
	answer, ok := <-p.pushBlock(data)
	if !ok {
		return nil, ErrNoAnswers
	}
	errMsg := answer.GetError()
	if errMsg != "" {
		return nil, errors.New(errMsg)
	}
	return answer.GetData(), nil
}

// GetSnapshot implements AppProxy interface method
func (p *GrpcAppProxy) GetSnapshot(blockIndex int64) ([]byte, error) {
	answer, ok := <-p.pushQuery(blockIndex)
	if !ok {
		return nil, ErrNoAnswers
	}
	errMsg := answer.GetError()
	if errMsg != "" {
		return nil, errors.New(errMsg)
	}
	return answer.GetData(), nil
}

// Restore implements AppProxy interface method
func (p *GrpcAppProxy) Restore(snapshot []byte) error {
	answer, ok := <-p.pushRestore(snapshot)
	if !ok {
		return ErrNoAnswers
	}
	errMsg := answer.GetError()
	if errMsg != "" {
		return errors.New(errMsg)
	}
	return nil
}

/*
 * staff:
 */

func (p *GrpcAppProxy) routeAnswer(hash *internal.ToServer_Answer) {
	uuid, err := xid.FromBytes(hash.GetUid())
	if err != nil {
		// TODO: log invalid uuid
		return
	}
	p.askingsSync.RLock()
	if ch, ok := p.askings[uuid]; ok {
		ch <- hash
	}
	p.askingsSync.RUnlock()
}

func (p *GrpcAppProxy) pushBlock(block []byte) chan *internal.ToServer_Answer {
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

func (p *GrpcAppProxy) pushQuery(index int64) chan *internal.ToServer_Answer {
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

func (p *GrpcAppProxy) pushRestore(snapshot []byte) chan *internal.ToServer_Answer {
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
	p.askingsSync.Lock()
	p.askings[uuid] = ch
	p.askingsSync.Unlock()
	// timeout
	go func() {
		<-time.After(p.timeout)
		p.askingsSync.Lock()
		delete(p.askings, uuid)
		p.askingsSync.Unlock()
		close(ch)
	}()

	return ch
}
