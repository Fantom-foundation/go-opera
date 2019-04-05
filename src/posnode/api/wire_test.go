package api

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"google.golang.org/grpc"

	"github.com/Fantom-foundation/go-lachesis/src/inter/wire"
	"github.com/Fantom-foundation/go-lachesis/src/posnode/network"
)

func TestGRPC(t *testing.T) {

	t.Run("over TCP", func(t *testing.T) {
		testGRPC(t, "", false)
	})

	t.Run("over Fake", func(t *testing.T) {
		dialer := network.FakeDialer("client.fake")
		testGRPC(t, "server.fake:55555", true, grpc.WithContextDialer(dialer))
	})
}

func testGRPC(t *testing.T, bind string, fake bool, opts ...grpc.DialOption) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// service mock
	svc := NewMockNodeServer(ctrl)
	svc.EXPECT().
		SyncEvents(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, req *KnownEvents) (*KnownEvents, error) {
			t.Logf("connection from '%s' host", GrpcPeerHost(ctx))
			return &KnownEvents{}, nil
		}).
		Times(1)
	svc.EXPECT().
		GetEvent(gomock.Any(), gomock.Any()).
		Return(&wire.Event{}, nil).
		Times(1)
	svc.EXPECT().
		GetPeerInfo(gomock.Any(), gomock.Any()).
		Return(&PeerInfo{}, nil).
		Times(1)

	// grpc server
	server, addr := StartService(bind, svc, t.Logf, fake)
	defer server.Stop()

	// grpc client
	conn, err := grpc.DialContext(context.Background(), addr, append(opts, grpc.WithInsecure())...)
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
