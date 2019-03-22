package wire

//NOTE: mockgen does not work properly out of GOPATH
//go:generate mockgen -package=wire -source=service.pb.go -destination=mock_test.go NodeServer

import (
	"context"
	"math"
	"net"
	"testing"

	"github.com/golang/mock/gomock"
	"google.golang.org/grpc"

	"github.com/Fantom-foundation/go-lachesis/src/posnode/network"
)

func TestGRPC(t *testing.T) {

	t.Run("over TCP", func(t *testing.T) {
		listener := network.TcpListener("")
		testGRPC(t, listener)
	})

	t.Run("over Fake", func(t *testing.T) {
		return
		listener := network.FakeListener("")
		testGRPC(t, listener, grpc.WithContextDialer(network.FakeDial))
	})
}

func testGRPC(t *testing.T, listener net.Listener, opts ...grpc.DialOption) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	srv := NewMockNodeServer(ctrl)
	srv.EXPECT().
		GetEvent(gomock.Any(), gomock.Any()).
		Return(&Event{}, nil).
		MinTimes(1)

	server := grpc.NewServer(
		grpc.MaxRecvMsgSize(math.MaxInt32),
		grpc.MaxSendMsgSize(math.MaxInt32))
	RegisterNodeServer(server, srv)

	go server.Serve(listener)
	defer server.Stop()

	netAddr := listener.Addr().String()
	t.Logf("connect address is %s", netAddr)
	conn, err := grpc.DialContext(context.Background(), netAddr, append(opts, grpc.WithInsecure())...)
	if err != nil {
		t.Fatal(err)
	}
	client := NewNodeClient(conn)

	_, err = client.GetEvent(context.Background(), &EventRequest{})
	if err != nil {
		t.Fatal(err)
	}
}
