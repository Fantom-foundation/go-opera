package proxy

import (
	"testing"
	"time"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

//go:generate mockgen -package=proxy -destination=mock_test.go github.com/Fantom-foundation/go-lachesis/src/proxy Node,Consensus

func TestCtrlAppProxy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	node := NewMockNode(ctrl)
	consensus := NewMockConsensus(ctrl)

	ctrlProxy, _ := NewGrpcCtrlProxy("localhost:55557", node, consensus, nil, nil)
	defer ctrlProxy.Close()

	cmdProxy, err := NewGrpcCmdProxy("localhost:55557", 100*time.Millisecond)
	if err != nil {
		t.Fatalf("prepare command proxy: %v", err)
	}

	id := "0x70210aeeb6f7550d1a3f0e6e1bd41fc9b7c6122b5176ed7d7fe93847dac856cf"
	peer := hash.HexToPeer(id)

	t.Run("id", func(t *testing.T) {
		assert := assert.New(t)

		node.EXPECT().GetID().Return(peer)
		got, err := cmdProxy.GetID()

		assert.NoError(err)
		assert.Equal(id, got)
	})

	t.Run("stake", func(t *testing.T) {
		assert := assert.New(t)

		node.EXPECT().GetID().Return(peer)
		consensus.EXPECT().GetStakeOf(peer).Return(0.0023)

		got, err := cmdProxy.GetStake()

		assert.NoError(err)
		assert.Equal(0.0023, got)
	})

	t.Run("internal_txn", func(t *testing.T) {
		assert := assert.New(t)

		tx := inter.InternalTransaction{
			Amount:   2,
			Receiver: peer,
		}
		node.EXPECT().AddInternalTxn(tx)

		err := cmdProxy.SubmitInternalTxn(2, id)

		assert.NoError(err)
	})
}
