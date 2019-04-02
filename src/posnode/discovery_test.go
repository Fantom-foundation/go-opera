package posnode

import (
	"math"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/posnode/network"
	"github.com/Fantom-foundation/go-lachesis/src/posnode/wire"
)

func Test_Node_AskPeerInfo(t *testing.T) {
	assert := assert.New(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	srv := wire.NewMockNodeServer(ctrl)

	server := grpc.NewServer(
		grpc.MaxRecvMsgSize(math.MaxInt32),
		grpc.MaxSendMsgSize(math.MaxInt32))
	wire.RegisterNodeServer(server, srv)

	listener := network.FakeListener("server.fake:55555")
	go server.Serve(listener)
	defer server.Stop()

	store := NewMemStore()
	n := NewForTests("any", store, nil)

	key, err := crypto.GenerateECDSAKey()
	if !assert.NoError(err) {
		return
	}
	id := CalcNodeID(&key.PublicKey)
	info := &wire.PeerInfo{
		ID:     id.Hex(),
		PubKey: common.FromECDSAPub(&key.PublicKey),
		Host:   "remote.server",
	}

	srv.EXPECT().
		GetPeerInfo(gomock.Any(), gomock.Any()).
		Return(info, nil)

	n.AskPeerInfo(hash.HexToPeer("known"), id, "server.fake")

	got := store.GetPeer(id)
	expect := WireToPeer(info)
	assert.Equal(expect, got, "AskPeerInfo()")

}
