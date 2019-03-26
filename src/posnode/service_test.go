package posnode

import (
	"context"
	reflect "reflect"
	"testing"
	"time"

	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	"github.com/Fantom-foundation/go-lachesis/src/posnode/wire"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Test_Node_GetPeerInfo(t *testing.T) {
	t.Log("with initialized node and cli")
	{
		store := NewMemStore()
		n := NewForTests("server.fake", store, nil)
		n.StartServiceForTests()
		defer n.StopService()

		// connect client to the node.
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		cli, err := n.ConnectTo(ctx, "server.fake:55555")
		if err != nil {
			t.Fatalf("failed to connect to node with gRPC: %v", err)
		}

		t.Log("\ttest:0\tshould return info about existng peer")
		{
			// initialize peer and insert it into the store.
			key, err := crypto.GenerateECDSAKey()
			if err != nil {
				t.Fatalf("failed to generate ecdsa key for peer: %v", err)
			}
			pubKey := key.PublicKey
			id := CalcNodeID(&pubKey)
			netAddr := "8.8.8.8:8083"

			peer := Peer{
				ID:      id,
				PubKey:  &pubKey,
				NetAddr: netAddr,
			}

			store.SetPeer(&peer)

			// make request
			in := wire.PeerRequest{
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
		}

		t.Log("\ttest:1\tshould return not found error")
		{
			in := wire.PeerRequest{
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
		}
	}
}
