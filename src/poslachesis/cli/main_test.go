package main

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/proxy"
)

//go:generate mockgen -package=main -destination=mock_test.go github.com/Fantom-foundation/go-lachesis/src/proxy Node,Consensus

func TestApp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	node := NewMockNode(ctrl)
	consensus := NewMockConsensus(ctrl)

	ctrlProxy, err := proxy.NewGrpcCtrlProxy("localhost:55557", node, consensus, nil, nil)
	if err != nil {
		t.Fatalf("failed to prepare ctrl proxy: %v", err)
	}
	defer ctrlProxy.Close()

	app := prepareApp()
	var out bytes.Buffer
	app.SetOutput(&out)

	id := "0x70210aeeb6f7550d1a3f0e6e1bd41fc9b7c6122b5176ed7d7fe93847dac856cf"
	peer := hash.HexToPeer(id)

	t.Run("id", func(t *testing.T) {
		assert := assert.New(t)
		node.EXPECT().GetID().Return(peer)

		app.SetArgs([]string{"id"})
		defer out.Reset()

		err := app.Execute()
		if !assert.NoError(err) {
			return
		}

		assert.Contains(out.String(), id)
	})

	t.Run("stake", func(t *testing.T) {
		assert := assert.New(t)

		node.EXPECT().GetID().Return(peer)
		consensus.EXPECT().GetStakeOf(peer).Return(0.0023)

		app.SetArgs([]string{"stake"})
		defer out.Reset()

		err := app.Execute()
		if !assert.NoError(err) {
			return
		}

		assert.Contains(out.String(), "0.0023")
	})

	t.Run("internal_txn missing flags", func(t *testing.T) {
		assert := assert.New(t)

		app.SetArgs([]string{"internal_txn"})
		defer out.Reset()

		err := app.Execute()
		if !assert.Error(err) {
			return
		}

		assert.Contains(out.String(), "required flag(s) \"amount\", \"receiver\" not set")
	})

	t.Run("internal_txn", func(t *testing.T) {
		assert := assert.New(t)

		tx := inter.InternalTransaction{
			Amount:   2,
			Receiver: peer,
		}
		node.EXPECT().AddInternalTxn(tx)

		app.SetArgs([]string{"internal_txn", "--amount=2", fmt.Sprintf("--receiver=%s", id)})
		defer out.Reset()

		err := app.Execute()
		if !assert.NoError(err) {
			return
		}

		assert.Contains(out.String(), "transaction has been added")
	})
}
