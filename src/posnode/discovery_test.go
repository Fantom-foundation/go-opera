package posnode

import (
	"math"
	"net"
	"reflect"
	"testing"

	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	"github.com/Fantom-foundation/go-lachesis/src/posnode/wire"

	"github.com/golang/mock/gomock"
	"google.golang.org/grpc"
)

func Test_Node_AskPeerInfo(t *testing.T) {
	t.Log("with initialized node")
	{
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		srv := NewMockNodeServer(ctrl)

		server := grpc.NewServer(grpc.MaxRecvMsgSize(math.MaxInt32), grpc.MaxSendMsgSize(math.MaxInt32))
		wire.RegisterNodeServer(server, srv)

		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatalf("failed to listen random port")
		}
		go server.Serve(listener)
		defer server.Stop()

		host, port, err := net.SplitHostPort(listener.Addr().String())
		if err != nil {
			t.Fatalf("failed to split host port: %s", listener.Addr())
		}

		store := NewMemStore()
		key, err := crypto.GenerateECDSAKey()
		if err != nil {
			t.Fatalf("failed to generate ecdsa key: %v", err)
		}

		n := NewWithName("node001", key, store, nil)
		n.defaultPort = port

		t.Log("should insert peer")
		{
			peerInfo := wire.PeerInfo{
				ID:      common.HexToAddress("unknown").Hex(),
				PubKey:  []byte{},
				NetAddr: "8.8.8.8:8083",
			}

			srv.EXPECT().GetPeerInfo(gomock.Any(), gomock.Any()).Return(&peerInfo, nil)
			n.AskPeerInfo(common.HexToAddress("known"), common.HexToAddress("unknown"), host)

			got, err := store.GetPeer(common.HexToAddress(peerInfo.ID))
			if err != nil {
				t.Fatalf("failed to get inserted peer: %v", err)
			}
			expect := WireToPeer(&peerInfo)

			if !reflect.DeepEqual(expect, got) {
				t.Errorf("expected to insert peer: %+v got: %+v", expect, got)
			}
		}
	}
}

func Test_netAddFromHostPort(t *testing.T) {
	tt := []struct {
		name       string
		host, port string
		expect     string
	}{
		{
			name:   "port 80",
			host:   "127.0.0.1",
			port:   "80",
			expect: "127.0.0.1:80",
		},
		{
			name:   "port :80",
			host:   "127.0.0.1",
			port:   ":80",
			expect: "127.0.0.1:80",
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := netAddFromHostPort(tc.host, tc.port)
			if got != tc.expect {
				t.Errorf("expected result to be: %s got: %s", tc.expect, got)
			}
		})
	}
}
