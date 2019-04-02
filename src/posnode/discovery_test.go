package posnode

import (
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"google.golang.org/grpc"

	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/posnode/api"
	"github.com/Fantom-foundation/go-lachesis/src/posnode/network"
)

func Test_Node_AskPeerInfo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	srv := api.NewMockNodeServer(ctrl)

	server := grpc.NewServer(grpc.MaxRecvMsgSize(math.MaxInt32), grpc.MaxSendMsgSize(math.MaxInt32))
	api.RegisterNodeServer(server, srv)

	listener := network.FakeListener("server.fake:55555")
	go server.Serve(listener)
	defer server.Stop()

	t.Run("happy path", func(t *testing.T) {
		store := NewMemStore()
		n := NewForTests("any", store, nil)

		idKey, err := crypto.GenerateECDSAKey()
		if err != nil {
			t.Fatalf("failed to generate key")
		}

		idp := CalcNodeID(&idKey.PublicKey)
		id := hash.HexToPeer(idp.Hex())

		peerInfo := api.PeerInfo{
			ID:     id.Hex(),
			PubKey: common.FromECDSAPub(&idKey.PublicKey),
			Host:   "remote.server:55555",
		}

		srv.EXPECT().GetPeerInfo(gomock.Any(), gomock.Any()).Return(&peerInfo, nil)

		sourceKey, err := crypto.GenerateECDSAKey()
		if err != nil {
			t.Fatalf("failed to generate key")
		}
		sp := CalcNodeID(&sourceKey.PublicKey)
		source := hash.HexToPeer(sp.Hex())

		n.AskPeerInfo(source, id, "server.fake")

		got := store.GetPeer(id)
		expect := WireToPeer(&peerInfo)
		if !reflect.DeepEqual(expect, got) {
			t.Errorf("expected to insert peer: %+v got: %+v", expect, got)
		}

		discovery := n.discovery.store.GetDiscovery(source)
		if discovery == nil {
			t.Error("expected to add discovery")
			return
		}

		if !discovery.Available {
			t.Error("expected to set discover availability as available")
		}
	})

	t.Run("should set unavailable", func(t *testing.T) {
		store := NewMemStore()
		n := NewForTests("any", store, nil)

		id := hash.HexToPeer("unknown")
		discovery := Discovery{
			ID:          id,
			Host:        "bad.server",
			LastRequest: time.Now().Truncate(time.Hour),
			Available:   false,
		}
		n.discovery.store.SetDiscovery(&discovery)

		source := hash.HexToPeer("known")
		n.AskPeerInfo(source, id, "bad.server")

		peer := store.GetPeer(id)
		if peer != nil {
			t.Error("sould not add peer")
		}

		newDiscovery := n.discovery.store.GetDiscovery(source)
		if newDiscovery.Available {
			t.Error("should set discovery availability as unavailable")
		}
	})

	t.Run("same id and source", func(t *testing.T) {
		store := NewMemStore()
		n := NewForTests("any", store, nil)

		idKey, err := crypto.GenerateECDSAKey()
		if err != nil {
			t.Fatalf("failed to generate key")
		}

		idp := CalcNodeID(&idKey.PublicKey)
		id := hash.HexToPeer(idp.Hex())

		peerInfo := api.PeerInfo{
			ID:     id.Hex(),
			PubKey: common.FromECDSAPub(&idKey.PublicKey),
			Host:   "remote.server:55555",
		}

		srv.EXPECT().GetPeerInfo(gomock.Any(), gomock.Any()).Return(&peerInfo, nil)
		source := id

		n.AskPeerInfo(source, id, "server.fake")

		got := store.GetPeer(id)
		expect := WireToPeer(&peerInfo)
		if !reflect.DeepEqual(expect, got) {
			t.Errorf("expected to insert peer: %+v got: %+v", expect, got)
		}

		discovery := n.discovery.store.GetDiscovery(source)
		if discovery == nil {
			t.Error("expected to add discovery")
			return
		}

		if !discovery.Available {
			t.Error("expected to set discover availability as available")
		}
	})
}

func Test_Node_CheckPeerIsKnown(t *testing.T) {
	store := NewMemStore()
	n := NewForTests("any", store, nil)
	n.discovery.tasks = make(chan discoveryTask, 1)

	t.Run("peer exists", func(t *testing.T) {
		idKey, err := crypto.GenerateECDSAKey()
		if err != nil {
			t.Fatalf("failed to generate key")
		}

		idp := CalcNodeID(&idKey.PublicKey)
		id := hash.HexToPeer(idp.Hex())

		peerInfo := api.PeerInfo{
			ID:     id.Hex(),
			PubKey: common.FromECDSAPub(&idKey.PublicKey),
			Host:   "remote.server:55555",
		}

		n.store.SetPeer(WireToPeer(&peerInfo))

		source := hash.HexToPeer("known")
		n.CheckPeerIsKnown(source, id, "bad.server")
		defer func() {
			n.discovery.store.Clear()
			n.discovery.tasks = make(chan discoveryTask, 1)
		}()

		if len(n.discovery.tasks) > 0 {
			t.Error("should not add task")
		}
	})

	t.Run("previous discovery", func(t *testing.T) {
		idKey, err := crypto.GenerateECDSAKey()
		if err != nil {
			t.Fatalf("failed to generate key")
		}

		idp := CalcNodeID(&idKey.PublicKey)
		id := hash.HexToPeer(idp.Hex())
		source := hash.HexToPeer("known")

		d := Discovery{
			ID:          source,
			Host:        "bad.server",
			Available:   false,
			LastRequest: time.Now(),
		}
		n.discovery.store.SetDiscovery(&d)

		n.CheckPeerIsKnown(source, id, "bad.server")
		defer func() {
			n.discovery.store.Clear()
			n.discovery.tasks = make(chan discoveryTask, 1)
		}()

		if len(n.discovery.tasks) > 0 {
			t.Error("should not add task")
		}
	})

	t.Run("happy path", func(t *testing.T) {
		id := hash.HexToPeer("source")
		source := hash.HexToPeer("known")

		n.CheckPeerIsKnown(source, id, "bad.server")
		defer func() {
			n.discovery.store.Clear()
			n.discovery.tasks = make(chan discoveryTask, 1)
		}()

		if len(n.discovery.tasks) != 1 {
			t.Error("should add task")
		}
	})
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
