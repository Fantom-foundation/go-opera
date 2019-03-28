package posnode

import (
	"math"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	//"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/posnode/network"
	"github.com/Fantom-foundation/go-lachesis/src/posnode/wire"
)

func Test_Node_AskPeerInfo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	srv := wire.NewMockNodeServer(ctrl)

	server := grpc.NewServer(grpc.MaxRecvMsgSize(math.MaxInt32), grpc.MaxSendMsgSize(math.MaxInt32))
	wire.RegisterNodeServer(server, srv)

	listener := network.FakeListener("server.fake:55555")
	go server.Serve(listener)
	defer server.Stop()

	store := NewMemStore()
	n := NewForTests("any", store, nil)

	peerInfo := wire.PeerInfo{
		ID:     common.HexToAddress("unknown").Hex(),
		PubKey: []byte{},
		Host:   "remote.server",
	}

	srv.EXPECT().
		GetPeerInfo(gomock.Any(), gomock.Any()).
		Return(&peerInfo, nil)

	n.AskPeerInfo(common.HexToAddress("known"), common.HexToAddress("unknown"), "server.fake")

	got := store.GetPeer(common.HexToAddress(peerInfo.ID))
	expect := WireToPeer(&peerInfo)

	if !reflect.DeepEqual(expect, got) {
		t.Errorf("expected to insert peer: %+v got: %+v", expect, got)
	}

}
