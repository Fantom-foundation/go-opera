package posnode

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	"github.com/Fantom-foundation/go-lachesis/src/posnode/api"
)

func Test_Node_GetPeerInfo(t *testing.T) {
	store := NewMemStore()
	n := NewForTests("server.fake", store, nil)

	n.StartServiceForTests()
	defer n.StopService()

	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	defer cancel()

	cli, err := n.ConnectTo(ctx, "server.fake")
	if !assert.NoError(t, err) {
		return
	}

	t.Run("existing peer", func(t *testing.T) {
		//assert := assert.New(t)
		// initialize peer and insert it into the store.
		key, err := crypto.GenerateECDSAKey()
		if err != nil {
			t.Fatalf("failed to generate ecdsa key for peer: %v", err)
		}
		pubKey := key.PublicKey
		id := CalcNodeID(&pubKey)
		netAddr := "8.8.8.8:8083"

		peer := Peer{
			ID:     id,
			PubKey: &pubKey,
			Host:   netAddr,
		}

		store.SetPeer(&peer)

		// make request
		in := api.PeerRequest{
			PeerID: id.Hex(),
		}
		got, err := cli.GetPeerInfo(ctx, &in)
		if err != nil {
			t.Fatalf("failed to make get peer info request: %v", err)
		}

		// check result
		expect := peer.ToWire()
		if !reflect.DeepEqual(expect, got) {
			t.Errorf("expected response to be: %+v, got: %+v", expect, got)
			return
		}
	})

	t.Run("no existing peer", func(t *testing.T) {
		in := api.PeerRequest{
			PeerID: "unknown",
		}
		_, err = cli.GetPeerInfo(ctx, &in)
		if err == nil {
			t.Error("expected server to return error")
		}
		s := status.Convert(err)

		if s.Code() != codes.NotFound {
			t.Errorf("expected return code to be: %d got: %d", codes.NotFound, s.Code())
		}
	})

}
