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
	serverID := hash.PeerOfPubkey(serverKey.Public())
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

	t.Run("authorized", func(t *testing.T) {
		assert := assert.New(t)

		SetGenesisHash(hash.FakeHash())

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
		id1, ctx1 := ServerPeerID(nil)
		_, err = client.SyncEvents(ctx1, &KnownEvents{})
		if !assert.NoError(err) {
			return
		}
		if !assert.Equal(serverID, *id1) {
			return
		}

		// GetEvent() rpc
		id2, ctx2 := ServerPeerID(nil)
		_, err = client.GetEvent(ctx2, &EventRequest{})
		if !assert.NoError(err) {
			return
		}
		if !assert.Equal(serverID, *id2) {
			return
		}

		// GetPeerInfo() rpc
		id3, ctx3 := ServerPeerID(nil)
		_, err = client.GetPeerInfo(ctx3, &PeerRequest{})
		if !assert.NoError(err) {
			return
		}
		if !assert.Equal(serverID, *id3) {
			return
		}
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
		id1, ctx1 := ServerPeerID(nil)
		_, err = client.SyncEvents(ctx1, &KnownEvents{})
		if !assert.Error(err) {
			return
		}
		if !assert.Equal(hash.EmptyPeer, *id1) {
			return
		}

		// GetEvent() rpc
		id2, ctx2 := ServerPeerID(nil)
		_, err = client.GetEvent(ctx2, &EventRequest{})
		if !assert.Error(err) {
			return
		}
		if !assert.Equal(hash.EmptyPeer, *id2) {
			return
		}

		// GetPeerInfo() rpc
		id3, ctx3 := ServerPeerID(nil)
		_, err = client.GetPeerInfo(ctx3, &PeerRequest{})
		if !assert.Error(err) {
			return
		}
		if !assert.Equal(hash.EmptyPeer, *id3) {
			return
		}
	})

	// TODO: test client with unauthorized server.
}
