package main

import (
	"bytes"
	"fmt"
	"math/rand"
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

	ctrlProxy, _, err := proxy.NewGrpcCtrlProxy("localhost:55557", node, consensus, nil, nil)
	if err != nil {
		t.Fatalf("failed to prepare ctrl proxy: %v", err)
	}
	defer ctrlProxy.Close()

	app := prepareApp()
	var out bytes.Buffer
	app.SetOutput(&out)

	peer := hash.FakePeer()

	t.Run("id", func(t *testing.T) {
		assert := assert.New(t)
		node.EXPECT().
			GetID().
			Return(peer)

		app.SetArgs([]string{"id"})
		defer out.Reset()

		err := app.Execute()
		if !assert.NoError(err) {
			return
		}

		assert.Contains(out.String(), peer.Hex())
	})

	t.Run("balance", func(t *testing.T) {
		assert := assert.New(t)

		amount := rand.Uint64()

		node.EXPECT().
			GetID().
			Return(peer)
		consensus.EXPECT().
			GetBalanceOf(peer).
			Return(amount)

		app.SetArgs([]string{"balance"})
		defer out.Reset()

		err := app.Execute()
		if !assert.NoError(err) {
			return
		}

		assert.Contains(out.String(), fmt.Sprint(amount))
	})

	t.Run("transfer missing flags", func(t *testing.T) {
		assert := assert.New(t)

		app.SetArgs([]string{"transfer"})
		defer out.Reset()

		err := app.Execute()
		if !assert.Error(err) {
			return
		}

		assert.Contains(out.String(), "required flag(s) \"amount\", \"receiver\" not set")
	})

	t.Run("transfer", func(t *testing.T) {
		assert := assert.New(t)

		tx := inter.InternalTransaction{
			Amount:   rand.Uint64(),
			Receiver: peer,
		}

		node.EXPECT().
			AddInternalTxn(tx)

		app.SetArgs([]string{
			"transfer",
			fmt.Sprintf("--amount=%d", tx.Amount),
			fmt.Sprintf("--receiver=%s", tx.Receiver.Hex())})
		defer out.Reset()

		err := app.Execute()
		if !assert.NoError(err) {
			return
		}

		assert.Contains(out.String(), "ok")
	})
}
