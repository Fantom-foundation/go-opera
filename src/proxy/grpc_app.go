package proxy

import (
	"errors"
	"io"
	"math"
	"net"
	"sync"
	"time"

	"github.com/rs/xid"
	"google.golang.org/grpc"

	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/logger"
	"github.com/Fantom-foundation/go-lachesis/src/network"
	"github.com/Fantom-foundation/go-lachesis/src/poset"
	"github.com/Fantom-foundation/go-lachesis/src/proxy/internal"
)

var errNoAnswers = errors.New("no answers")

type (
	// appStream  a shortcut for generated type.
	appStream internal.Lachesis_ConnectServer

	// grpcAppProxy implements the AppProxy interface.
	grpcAppProxy struct {
		listener net.Listener
		server   *grpc.Server

		timeout     time.Duration
		newClients  chan appStream
		askings     map[xid.ID]chan *internal.ToServer_Answer
		askingsSync sync.RWMutex

		event4server  chan []byte
		event4clients chan *internal.ToClient

		shutdown chan struct{}
		wg       sync.WaitGroup

		logger.Instance
	}
)

// NewGrpcAppProxy instantiates a joined AppProxy-interface listen to remote apps.
func NewGrpcAppProxy(
	bind string, timeout time.Duration, listen network.ListenFunc,
) (
	res AppProxy, addr string, err error,
) {
	if listen == nil {
		listen = network.TCPListener
	}

	p := &grpcAppProxy{
		timeout:    timeout,
		newClients: make(chan appStream, 100),
		// TODO: buffer channels?
		askings:       make(map[xid.ID]chan *internal.ToServer_Answer),
		event4server:  make(chan []byte, 5),
		event4clients: make(chan *internal.ToClient),

		shutdown: make(chan struct{}),

		Instance: logger.MakeInstance(),
	}

	p.listener = listen(bind)

	p.server = grpc.NewServer(
		grpc.MaxRecvMsgSize(math.MaxInt32),
		grpc.MaxSendMsgSize(math.MaxInt32))
	internal.RegisterLachesisServer(p.server, p)

	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		if err := p.server.Serve(p.listener); err != nil {
			p.Fatal(err)
		}
	}()

	p.wg.Add(1)
	go p.sendEvents4clients()

	return p, p.listener.Addr().String(), nil
}

func (p *grpcAppProxy) Close() {
	p.server.Stop()
	close(p.shutdown)
	p.wg.Wait()
	close(p.event4server)
}

/*
 * network interface:
 */

// Connect implements gRPC-server interface: LachesisServer.
func (p *grpcAppProxy) Connect(stream internal.Lachesis_ConnectServer) error {
	p.wg.Add(1)
	defer p.wg.Done()
	// save client's stream for writing
	p.newClients <- stream
	p.Debugf("client connected")
	// read from stream
	for {
		req, err := stream.Recv()
		if err != nil {
			if err != io.EOF {
				p.Debugf("client refused: %s", err)
			} else {
				p.Debugf("client disconnected well")
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

func (p *grpcAppProxy) sendEvents4clients() {
	defer p.wg.Done()
	defer close(p.event4clients)

	var (
		err       error
		connected []appStream
		alive     []appStream
		stream    appStream
	)

	eventProcessFunc := func(event *internal.ToClient) {
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

until_shutdown:
	for {
		select {
		case event := <-p.event4clients:
			eventProcessFunc(event)
		case <-p.shutdown:
			break until_shutdown
		}
	}

until_empty:
	for {
		select {
		case event := <-p.event4clients:
			eventProcessFunc(event)
		default:
			break until_empty
		}
	}
}

/*
 * inmem interface: AppProxy implementation
 */

// SubmitCh implements AppProxy interface method.
func (p *grpcAppProxy) SubmitCh() chan []byte {
	return p.event4server
}

// SubmitInternalCh implements AppProxy interface method.
// TODO: Incorrect implementation, just adding to the interface so long.
func (p *grpcAppProxy) SubmitInternalCh() chan inter.InternalTransaction {
	return nil
}

// CommitBlock implements AppProxy interface method.
func (p *grpcAppProxy) CommitBlock(block poset.Block) ([]byte, error) {
	data, err := block.ProtoMarshal()
	if err != nil {
		return nil, err
	}
	answer, ok := <-p.pushBlock(data)
	if !ok {
		return nil, errNoAnswers
	}
	errMsg := answer.GetError()
	if errMsg != "" {
		return nil, errors.New(errMsg)
	}
	return answer.GetData(), nil
}

// GetSnapshot implements AppProxy interface method.
func (p *grpcAppProxy) GetSnapshot(blockIndex int64) ([]byte, error) {
	answer, ok := <-p.pushQuery(blockIndex)
	if !ok {
		return nil, errNoAnswers
	}
	errMsg := answer.GetError()
	if errMsg != "" {
		return nil, errors.New(errMsg)
	}
	return answer.GetData(), nil
}

// Restore implements AppProxy interface method.
func (p *grpcAppProxy) Restore(snapshot []byte) error {
	answer, ok := <-p.pushRestore(snapshot)
	if !ok {
		return errNoAnswers
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

func (p *grpcAppProxy) routeAnswer(hash *internal.ToServer_Answer) {
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

func (p *grpcAppProxy) pushBlock(block []byte) chan *internal.ToServer_Answer {
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

func (p *grpcAppProxy) pushQuery(index int64) chan *internal.ToServer_Answer {
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

func (p *grpcAppProxy) pushRestore(snapshot []byte) chan *internal.ToServer_Answer {
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

func (p *grpcAppProxy) subscribe4answer(uuid xid.ID) chan *internal.ToServer_Answer {
	ch := make(chan *internal.ToServer_Answer)

	p.askingsSync.Lock()
	p.askings[uuid] = ch
	p.askingsSync.Unlock()

	// timeout
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()

		<-time.After(p.timeout)
		p.askingsSync.Lock()
		delete(p.askings, uuid)
		p.askingsSync.Unlock()
		close(ch)
	}()

	return ch
}
