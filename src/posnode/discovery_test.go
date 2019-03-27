package posnode

import (
	"math"

	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"google.golang.org/grpc"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/posnode/api"
	"github.com/Fantom-foundation/go-lachesis/src/posnode/network"
)

func Test_Node_AskPeerInfo(t *testing.T) {
	t.Log("with initialized node")
	{
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		srv := api.NewMockNodeServer(ctrl)

		server := grpc.NewServer(grpc.MaxRecvMsgSize(math.MaxInt32), grpc.MaxSendMsgSize(math.MaxInt32))
		api.RegisterNodeServer(server, srv)

		listener := network.FakeListener("server.fake:55555")
		go server.Serve(listener)
		defer server.Stop()

		t.Log("\ttest:0\tshould add new peer and discovery")
		{
			store := NewMemStore()
			n := NewForTests("any", store, nil)

			id := hash.HexToPeer("unknown")
			peerInfo := api.PeerInfo{
				ID:     id.Hex(),
				PubKey: []byte{},
				Host:   "remote.server:55555",
			}

			srv.EXPECT().GetPeerInfo(gomock.Any(), gomock.Any()).Return(&peerInfo, nil)
			source := hash.HexToPeer("known")
			n.AskPeerInfo(source, id, "server.fake")

			got := store.GetPeer(id)
			expect := WireToPeer(&peerInfo)
			if !reflect.DeepEqual(expect, got) {
				t.Errorf("expected to insert peer: %+v got: %+v", expect, got)
			}

			discovery := store.GetDiscovery(source)
			if discovery == nil {
				t.Error("expected to add discovery")
			}

			if !discovery.Available {
				t.Error("expected to set discover availability as available")
			}
		}

		t.Log("\ttest:1\tshould set discovery availability as unavailable")
		{
			store := NewMemStore()
			n := NewForTests("any", store, nil)

			id := hash.HexToPeer("unknown")
			discovery := Discovery{
				ID:          id,
				Host:        "bad.server",
				LastRequest: time.Now().Truncate(time.Hour),
				Available:   false,
			}
			store.SetDiscovery(&discovery)

			source := hash.HexToPeer("known")
			n.AskPeerInfo(source, id, "bad.server")

			peer := store.GetPeer(id)
			if peer != nil {
				t.Error("sould not add peer")
			}

			newDiscovery := store.GetDiscovery(source)
			if newDiscovery.Available {
				t.Error("should set discovery availability as unavailable")
			}
		}
	}
}

func Test_shouldSkipDiscovery(t *testing.T) {
	tt := []struct {
		name      string
		discovery *Discovery
		host      string
		expect    bool
	}{
		{
			name:      "discovery is nil",
			discovery: nil,
			host:      "any",
			expect:    false,
		},
		{
			name: "available",
			discovery: &Discovery{
				Available: true,
			},
			host:   "any",
			expect: false,
		},
		{
			name: "new host",
			discovery: &Discovery{
				Host:      "old",
				Available: false,
			},
			host:   "new",
			expect: false,
		},
		{
			name: "wait time passed",
			discovery: &Discovery{
				Host:        "same",
				Available:   false,
				LastRequest: time.Now().Truncate(time.Hour),
			},
			host:   "same",
			expect: false,
		},
		{
			name: "happy path",
			discovery: &Discovery{
				Host:        "same",
				Available:   false,
				LastRequest: time.Now(),
			},
			host:   "same",
			expect: true,
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := shouldSkipDiscovery(tc.discovery, tc.host)
			if got != tc.expect {
				t.Errorf("expected: %v got: %v", tc.expect, got)
			}
		})
	}
}
