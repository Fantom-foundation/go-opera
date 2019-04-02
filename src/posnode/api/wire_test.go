package api

import (
	"context"
	"math"
	"net"
	"testing"

	"github.com/golang/mock/gomock"
	"google.golang.org/grpc"

	"github.com/Fantom-foundation/go-lachesis/src/inter/wire"
	"github.com/Fantom-foundation/go-lachesis/src/posnode/network"
)

func TestGRPC(t *testing.T) {

	t.Run("over TCP", func(t *testing.T) {
		listener := network.TcpListener("")
		testGRPC(t, listener)
	})

	t.Run("over Fake", func(t *testing.T) {
		listener := network.FakeListener("server.fake:55555")
		dialer := network.FakeDialer("client.fake")
		testGRPC(t, listener, grpc.WithContextDialer(dialer))
	})
}

func testGRPC(t *testing.T, listener net.Listener, opts ...grpc.DialOption) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// service mock
	srv := NewMockNodeServer(ctrl)
	srv.EXPECT().
		SyncEvents(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, req *KnownEvents) (*KnownEvents, error) {
			t.Logf("connection from '%s' host", GrpcPeerHost(ctx))
			return &KnownEvents{}, nil
		}).
		Times(1)
	srv.EXPECT().
		GetEvent(gomock.Any(), gomock.Any()).
		Return(&wire.Event{}, nil).
		Times(1)
	srv.EXPECT().
		GetPeerInfo(gomock.Any(), gomock.Any()).
		Return(&PeerInfo{}, nil).
		Times(1)

	// grpc server
	server := grpc.NewServer(
		grpc.MaxRecvMsgSize(math.MaxInt32),
		grpc.MaxSendMsgSize(math.MaxInt32))
	RegisterNodeServer(server, srv)

	go server.Serve(listener)
	defer server.Stop()

	// grpc client
	netAddr := listener.Addr().String()
	t.Logf("listen at '%s'", netAddr)
	conn, err := grpc.DialContext(context.Background(), netAddr, append(opts, grpc.WithInsecure())...)
	if err != nil {
		t.Fatal(err)
	}
	client := NewNodeClient(conn)

	// SyncEvents() rpc
	_, err = client.SyncEvents(context.Background(), &KnownEvents{})
	if err != nil {
		t.Fatal(err)
	}

	// GetEvent() rpc
	_, err = client.GetEvent(context.Background(), &EventRequest{})
	if err != nil {
		t.Fatal(err)
	}

	// GetPeerInfo() rpc
	_, err = client.GetPeerInfo(context.Background(), &PeerRequest{})
	if err != nil {
		t.Fatal(err)
	}
}
