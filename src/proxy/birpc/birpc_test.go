package birpc_test

import (
	"math/rand"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
	"sync"
	"testing"
	"time"

	"github.com/andrecronje/lachesis/src/proxy/birpc"
	ws "github.com/gorilla/websocket"
)

const wsPort = "2134"
const numRequests = 100

type RpcCalls struct{}

func (r *RpcCalls) Double(param int, reply *int) error {
	*reply = (param) * 2
	return nil
}

func (r *RpcCalls) Bytes(numBytes int, bytes *[]byte) error {
	b := make([]byte, numBytes)
	for i := 0; i < numBytes; i++ {
		b[i] = byte(i)
	}
	*bytes = b
	return nil
}

func TestBiRPC(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(numRequests * 2)

	// Setup WS server
	wsUpgrade := func(w http.ResponseWriter, r *http.Request) {
		upgrader := ws.Upgrader{}
		wsConn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatal(err)
		}

		birpc := birpc.New(wsConn)
		s := rpc.NewServer()
		err = s.Register(new(RpcCalls))
		if err != nil {
			t.Fatal(err)
		}

		go s.ServeCodec(jsonrpc.NewServerCodec(&birpc.Server))
		rpcClient := jsonrpc.NewClient(&birpc.Client)

		for i := 0; i < numRequests; i++ {
			go func(i int) {
				arg := rand.Intn(100)
				if i%2 == 0 {
					var reply int
					err := rpcClient.Call("RpcCalls.Double", arg, &reply)
					if err != nil {
						t.Fatal(err)
					}
					t.Log("RPC Call from server: sent:", arg, "got:", reply)
					if reply != arg*2 {
						t.Fatal("RPC Call from server: sent:", arg, "got:", reply, "expected:", arg*2)
					}
				} else {
					var reply []byte
					err := rpcClient.Call("RpcCalls.Bytes", arg, &reply)
					if err != nil {
						t.Fatal(err)
					}
					t.Log("RPC Call from server: sent:", arg, "got:", reply)
					for i, b := range reply {
						if int(b) != i {
							t.Fatal("RPC Call from server: sent:", arg, "got:", reply, "expected: correct array")
						}
					}
				}
				wg.Done()
			}(i)
		}
	}
	go http.ListenAndServe(":"+wsPort, http.HandlerFunc(wsUpgrade))
	time.Sleep(time.Millisecond * 5) // wait for http server to start

	// Setup WS client
	dialer := ws.DefaultDialer
	dialer.HandshakeTimeout = time.Second
	wsConn, _, err := dialer.Dial("ws://127.0.0.1:"+wsPort, nil)
	if err != nil {
		t.Fatal(err)
	}

	birpc := birpc.New(wsConn)
	s := rpc.NewServer()
	err = s.Register(new(RpcCalls))
	if err != nil {
		t.Fatal(err)
	}

	go s.ServeCodec(jsonrpc.NewServerCodec(&birpc.Server))
	rpcClient := jsonrpc.NewClient(&birpc.Client)

	for i := 0; i < numRequests; i++ {
		go func(i int) {
			arg := rand.Intn(100)
			if i%2 != 0 {
				var reply int
				err := rpcClient.Call("RpcCalls.Double", arg, &reply)
				if err != nil {
					t.Fatal(err)
				}
				t.Log("RPC Call from client: sent:", arg, "got:", reply)
				if reply != arg*2 {
					t.Fatal("RPC Call from client: sent:", arg, "got:", reply, "expected:", arg*2)
				}
			} else {
				var reply []byte
				err := rpcClient.Call("RpcCalls.Bytes", arg, &reply)
				if err != nil {
					t.Fatal(err)
				}
				t.Log("RPC Call from client: sent:", arg, "got:", reply)
				for i, b := range reply {
					if int(b) != i {
						t.Fatal("RPC Call from client: sent:", arg, "got:", reply, "expected: correct array")
					}
				}
			}
			wg.Done()
		}(i)
	}

	wg.Wait()
}
