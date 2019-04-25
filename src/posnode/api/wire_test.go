package api

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter/wire"
	"github.com/Fantom-foundation/go-lachesis/src/network"
)

func TestGRPC(t *testing.T) {

	t.Run("over TCP", func(t *testing.T) {
		testGRPC(t, "", "::1", network.TCPListener)
	})

	t.Run("over Fake", func(t *testing.T) {
		from := "client.fake"
		dialer := network.FakeDialer(from)
		testGRPC(t, "server.fake:0", from, network.FakeListener, grpc.WithContextDialer(dialer))
	})
}

func testGRPC(t *testing.T, bind, from string, listen network.ListenFunc, opts ...grpc.DialOption) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// keys
	serverKey := crypto.GenerateKey()
	//serverID := hash.PeerOfPubkey(serverKey.Public())
	clientKey := crypto.GenerateKey()
	clientID := hash.PeerOfPubkey(clientKey.Public())

	// service
	svc := NewMockNodeServer(ctrl)
	svc.EXPECT().
		SyncEvents(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, req *KnownEvents) (*KnownEvents, error) {
			assert.Equal(t, from, GrpcPeerHost(ctx))
			assert.Equal(t, clientID, GrpcPeerID(ctx))
			return &KnownEvents{}, nil
		}).
		AnyTimes()
	svc.EXPECT().
		GetEvent(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, req *EventRequest) (*wire.Event, error) {
			assert.Equal(t, from, GrpcPeerHost(ctx))
			assert.Equal(t, clientID, GrpcPeerID(ctx))
			return &wire.Event{}, nil
		}).
		AnyTimes()
	svc.EXPECT().
		GetPeerInfo(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, req *PeerRequest) (*PeerInfo, error) {
			assert.Equal(t, from, GrpcPeerHost(ctx))
			assert.Equal(t, clientID, GrpcPeerID(ctx))
			return &PeerInfo{}, nil
		}).
		AnyTimes()

	// server
	server, addr := StartService(bind, serverKey, svc, t.Logf, listen)
	defer server.Stop()

	t.Run("authorized client", func(t *testing.T) {
		assert := assert.New(t)

		opts := append(opts,
			grpc.WithInsecure(),
			grpc.WithUnaryInterceptor(ClientAuth(clientKey)),
		)
		conn, err := grpc.DialContext(context.Background(), addr, opts...)
		if err != nil {
			t.Fatal(err)
		}
		client := NewNodeClient(conn)

		// SyncEvents() rpc
		_, err = client.SyncEvents(context.Background(), &KnownEvents{})
		if !assert.NoError(err) {
			return
		}
		// TODO: got peer ID and compare with serverID

		// GetEvent() rpc
		_, err = client.GetEvent(context.Background(), &EventRequest{})
		if !assert.NoError(err) {
			return
		}
		// TODO: got peer ID and compare with serverID

		// GetPeerInfo() rpc
		_, err = client.GetPeerInfo(context.Background(), &PeerRequest{})
		if !assert.NoError(err) {
			return
		}
		// TODO: got peer ID and compare with serverID
	})

	t.Run("unauthorized client", func(t *testing.T) {
		assert := assert.New(t)

		opts := append(opts,
			grpc.WithInsecure(),
		)
		conn, err := grpc.DialContext(context.Background(), addr, opts...)
		if err != nil {
			t.Fatal(err)
		}
		client := NewNodeClient(conn)

		// SyncEvents() rpc
		_, err = client.SyncEvents(context.Background(), &KnownEvents{})
		if !assert.Error(err) {
			return
		}

		// GetEvent() rpc
		_, err = client.GetEvent(context.Background(), &EventRequest{})
		if !assert.Error(err) {
			return
		}

		// GetPeerInfo() rpc
		_, err = client.GetPeerInfo(context.Background(), &PeerRequest{})
		if !assert.Error(err) {
			return
		}
	})

	// TODO: test client with unauthorized server.
}
