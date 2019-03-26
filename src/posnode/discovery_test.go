package posnode

import (
	"math"
	"reflect"
	"testing"

	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/posnode/network"
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

		listener := network.FakeListener("server.fake:55555")
		go server.Serve(listener)
		defer server.Stop()

		store := NewMemStore()
		n := NewForTests("any", store, nil)

		t.Log("\ttest:0\tshould discover new peer")
		{
			peerInfo := wire.PeerInfo{
				ID:      common.HexToAddress("unknown").Hex(),
				PubKey:  []byte{},
				NetAddr: "remote.server:55555",
			}

			srv.EXPECT().GetPeerInfo(gomock.Any(), gomock.Any()).Return(&peerInfo, nil)
			n.AskPeerInfo(common.HexToAddress("known"), common.HexToAddress("unknown"), "server.fake")

			got := store.GetPeer(common.HexToAddress(peerInfo.ID))
			expect := WireToPeer(&peerInfo)

			if !reflect.DeepEqual(expect, got) {
				t.Errorf("expected to insert peer: %+v got: %+v", expect, got)
			}
		}
	}
}
