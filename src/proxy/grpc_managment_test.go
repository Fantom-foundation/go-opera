package proxy

import (
	"context"
	"testing"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/proxy/wire"
	gomock "github.com/golang/mock/gomock"
	empty "github.com/golang/protobuf/ptypes/empty"
	"github.com/stretchr/testify/assert"
)

//go:generate mockgen -package=proxy-source=./grpc_management.go -destination=mock_test.go

func TestManagementServer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	node := NewMockNode(ctrl)
	consensus := NewMockConsensus(ctrl)

	server := NewManagementServer("localhost:55557", node, consensus, nil)
	defer server.Stop()

	client, err := NewManagementClient("localhost:55557")
	if err != nil {
		t.Fatalf("connect to server: %v", err)
	}

	id := "0x70210aeeb6f7550d1a3f0e6e1bd41fc9b7c6122b5176ed7d7fe93847dac856cf"
	peer := hash.HexToPeer(id)

	t.Run("id", func(t *testing.T) {
		assert := assert.New(t)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		node.EXPECT().GetID().Return(peer)
		resp, err := client.ID(ctx, &empty.Empty{})

		assert.NoError(err)
		assert.Equal(id, resp.Id)
	})

	t.Run("stake", func(t *testing.T) {
		assert := assert.New(t)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		node.EXPECT().GetID().Return(peer)
		consensus.EXPECT().GetStakeOf(peer).Return(0.0023)

		resp, err := client.Stake(ctx, &empty.Empty{})

		assert.NoError(err)
		assert.Equal(0.0023, resp.Value)
	})

	t.Run("internal_txn", func(t *testing.T) {
		assert := assert.New(t)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		tx := inter.InternalTransaction{
			Amount:   2,
			Receiver: peer,
		}
		node.EXPECT().AddInternalTxn(tx)

		_, err := client.InternalTxn(ctx, &wire.InternalTxnRequest{
			Amount:   2,
			Receiver: id,
		})

		assert.NoError(err)
	})
}
