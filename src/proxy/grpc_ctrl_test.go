package proxy

import (
	"math/rand"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/network"
)

func TestGrpcCtrlCalls(t *testing.T) {
	t.Run("over TCP", func(t *testing.T) {
		testGrpcCtrlCalls(t, network.TCPListener)
	})

	t.Run("over Fake", func(t *testing.T) {
		dialer := network.FakeDialer("client.fake")
		testGrpcCtrlCalls(t, network.FakeListener, grpc.WithContextDialer(dialer))
	})
}

func testGrpcCtrlCalls(t *testing.T, listen network.ListenFunc, opts ...grpc.DialOption) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	node := NewMockNode(ctrl)
	consensus := NewMockConsensus(ctrl)
	peer := hash.FakePeer()

	s, addr, err := NewGrpcCtrlProxy("127.0.0.1:", node, consensus, nil, nil)
	if !assert.NoError(t, err) {
		return
	}
	defer s.Close()

	c, err := NewGrpcNodeProxy(addr, nil)
	if !assert.NoError(t, err) {
		return
	}
	defer c.Close()

	t.Run("id", func(t *testing.T) {
		assert := assert.New(t)

		node.EXPECT().
			GetID().
			Return(peer).
			Times(1)

		got, err := c.GetSelfID()

		assert.NoError(err)
		assert.Equal(peer, got)
	})

	t.Run("stake", func(t *testing.T) {
		assert := assert.New(t)

		amount := rand.Uint64()

		consensus.EXPECT().
			GetBalanceOf(peer).
			Return(amount)

		got, err := c.GetBalanceOf(peer)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(amount, got)
	})

	t.Run("internal_txn", func(t *testing.T) {
		assert := assert.New(t)

		tx := inter.InternalTransaction{
			Amount:   rand.Uint64(),
			Receiver: peer,
		}
		node.EXPECT().
			AddInternalTxn(tx)

		err := c.SendTo(tx.Amount, tx.Receiver)
		assert.NoError(err)
	})
}
