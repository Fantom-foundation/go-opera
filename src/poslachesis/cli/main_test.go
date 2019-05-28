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
			StakeOf(peer).
			Return(amount)

		app.SetArgs([]string{"balance"})
		defer out.Reset()

		err := app.Execute()
		if !assert.NoError(err) {
			return
		}

		expect := fmt.Sprintf("balance of %s == %d", peer.Hex(), amount)
		assert.Contains(out.String(), expect)
	})

	t.Run("balance with peer flag", func(t *testing.T) {
		assert := assert.New(t)

		amount := rand.Uint64()

		otherPeer := hash.FakePeer()

		consensus.EXPECT().
			StakeOf(otherPeer).
			Return(amount)

		app.SetArgs([]string{
			"balance",
			fmt.Sprintf("--peer=%s", otherPeer.Hex())})
		defer out.Reset()

		err := app.Execute()
		if !assert.NoError(err) {
			return
		}

		expect := fmt.Sprintf("balance of %s == %d", otherPeer.Hex(), amount)
		assert.Contains(out.String(), expect)
	})

	t.Run("info not found", func(t *testing.T) {
		assert := assert.New(t)

		hex := "0x00000"
		consensus.EXPECT().
			GetTransaction(hash.HexToTransactionHash(hex)).
			Return(nil)

		app.SetArgs([]string{
			"info",
			hex,
		})
		defer out.Reset()

		err := app.Execute()
		if !assert.Error(err) {
			return
		}

		assert.Contains(out.String(), "transaction not found")
	})

	t.Run("info ok", func(t *testing.T) {
		assert := assert.New(t)

		h := hash.FakeTransaction()
		amount := rand.Uint64()

		tx := inter.InternalTransaction{
			Index:    1,
			Amount:   amount,
			Receiver: peer,
		}

		consensus.EXPECT().
			GetTransaction(h).
			Return(&tx)

		app.SetArgs([]string{
			"info",
			h.Hex(),
		})
		defer out.Reset()

		err := app.Execute()
		if !assert.NoError(err) {
			return
		}

		assert.Contains(
			out.String(),
			fmt.Sprintf(
				"transfer %d to %s",
				tx.Amount,
				tx.Receiver.Hex(),
			),
		)
	})

	t.Run("transfer missing flags", func(t *testing.T) {
		assert := assert.New(t)

		app.SetArgs([]string{"transfer"})
		defer out.Reset()

		err := app.Execute()
		if !assert.Error(err) {
			return
		}

		assert.Contains(out.String(), "required flag(s) \"amount\", \"index\", \"receiver\" not set")
	})

	t.Run("transfer", func(t *testing.T) {
		assert := assert.New(t)

		amount := rand.Uint64()

		h := hash.FakeTransaction()
		tx := inter.InternalTransaction{
			Index:    1,
			Amount:   amount,
			Receiver: peer,
		}

		node.EXPECT().
			AddInternalTxn(gomock.Any()).
			Return(h, nil)

		app.SetArgs([]string{
			"transfer",
			fmt.Sprintf("--index=%d", tx.Index),
			fmt.Sprintf("--amount=%d", tx.Amount),
			fmt.Sprintf("--receiver=%s", tx.Receiver.Hex())})
		defer out.Reset()

		err := app.Execute()
		if !assert.NoError(err) {
			return
		}

		assert.Contains(out.String(), h.Hex())
	})

}
