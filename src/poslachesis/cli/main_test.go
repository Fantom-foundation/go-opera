//go:generate mockgen -package=main -destination=mock_test.go github.com/Fantom-foundation/go-lachesis/src/proxy Node,Consensus
package main

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/proxy"
)

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
	reset := func() {
		сlearFlags(app)
		out.Reset()
	}

	peer := hash.FakePeer()

	t.Run("id", func(t *testing.T) {
		assertar := assert.New(t)
		reset()

		node.EXPECT().
			GetID().
			Return(peer)

		app.SetArgs([]string{"id"})

		err := app.Execute()
		if !assertar.NoError(err) {
			return
		}

		assertar.Contains(out.String(), peer.Hex())
	})

	t.Run("stake", func(t *testing.T) {
		assertar := assert.New(t)
		reset()

		amount := inter.Stake(rand.Uint64())

		node.EXPECT().
			GetID().
			Return(peer)
		consensus.EXPECT().
			StakeOf(peer).
			Return(amount)

		app.SetArgs([]string{"stake"})

		err := app.Execute()
		if !assertar.NoError(err) {
			return
		}

		expect := fmt.Sprintf("stake of %s == %d", peer.Hex(), amount)
		assertar.Contains(out.String(), expect)
	})

	t.Run("balance with peer flag", func(t *testing.T) {
		assertar := assert.New(t)
		reset()

		amount := inter.Stake(rand.Uint64())

		otherPeer := hash.FakePeer()

		consensus.EXPECT().
			StakeOf(otherPeer).
			Return(amount)

		app.SetArgs([]string{
			"stake",
			fmt.Sprintf("--peer=%s", otherPeer.Hex())})

		err := app.Execute()
		if !assertar.NoError(err) {
			return
		}

		expect := fmt.Sprintf("stake of %s == %d", otherPeer.Hex(), amount)
		assertar.Contains(out.String(), expect)
	})

	t.Run("txn not found", func(t *testing.T) {
		assertar := assert.New(t)
		reset()

		h := hash.FakeTransaction()
		node.EXPECT().
			GetInternalTxn(h).
			Return(nil, nil)

		app.SetArgs([]string{
			"txn",
			h.Hex(),
		})

		err := app.Execute()
		if !assertar.Error(err) {
			return
		}

		assertar.Contains(out.String(), "transaction not found")
	})

	t.Run("txn found", func(t *testing.T) {
		assertar := assert.New(t)
		reset()

		h := hash.FakeTransaction()
		amount := inter.Stake(rand.Uint64())

		txn := &inter.InternalTransaction{
			Nonce:    1,
			Amount:   amount,
			Receiver: peer,
		}

		node.EXPECT().
			GetInternalTxn(h).
			Return(txn, nil)

		app.SetArgs([]string{
			"txn",
			h.Hex(),
		})

		err := app.Execute()
		if !assertar.NoError(err) {
			return
		}

		assertar.Contains(
			out.String(),
			fmt.Sprintf(
				"transfer %d to %s",
				txn.Amount,
				txn.Receiver.Hex(),
			),
		)
	})

	t.Run("transfer missing flags", func(t *testing.T) {
		assertar := assert.New(t)
		reset()

		app.SetArgs([]string{"transfer"})

		err := app.Execute()
		if !assertar.Error(err) {
			return
		}

		assertar.Contains(out.String(), "required flag(s) \"amount\", \"index\", \"receiver\" not set")
	})

	t.Run("transfer", func(t *testing.T) {
		assertar := assert.New(t)
		reset()

		amount := inter.Stake(rand.Uint64())

		h := hash.FakeTransaction()
		tx := inter.InternalTransaction{
			Nonce:    1,
			Amount:   amount,
			Receiver: peer,
		}

		node.EXPECT().
			AddInternalTxn(gomock.Any()).
			Return(h, nil)

		app.SetArgs([]string{
			"transfer",
			fmt.Sprintf("--index=%d", tx.Nonce),
			fmt.Sprintf("--amount=%d", tx.Amount),
			fmt.Sprintf("--receiver=%s", tx.Receiver.Hex())})

		err := app.Execute()
		if !assertar.NoError(err) {
			return
		}

		assertar.Contains(out.String(), h.Hex())
	})

	t.Run("log-level one argument", func(t *testing.T) {
		assertar := assert.New(t)
		reset()

		app.SetArgs([]string{
			"log-level",
		})

		err := app.Execute()
		if !assertar.Error(err) {
			return
		}

		assertar.Contains(out.String(), "expected exactly one argument")
	})

	t.Run("log-level ok", func(t *testing.T) {
		assertar := assert.New(t)
		reset()

		app.SetArgs([]string{
			"log-level",
			"info",
		})

		err := app.Execute()
		if !assertar.NoError(err) {
			return
		}

		assertar.Contains(out.String(), "ok")
	})

	t.Run("key ok", func(t *testing.T) {
		assertar := assert.New(t)
		reset()

		app.SetArgs([]string{
			"key",
		})

		err := app.Execute()
		if !assertar.NoError(err) {
			return
		}

		err = os.Remove("./priv_key.pem")
		if !assertar.NoError(err) {
			return
		}

		assertar.Contains(out.String(), "priv_key.pem created")
	})
}

func сlearFlags(cmd *cobra.Command) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if err := f.Value.Set(f.DefValue); err != nil {
			panic(err)
		}
		f.Changed = false
	})

	for _, sub := range cmd.Commands() {
		сlearFlags(sub)
	}
}
