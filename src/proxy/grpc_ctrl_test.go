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

	t.Run("get self id", func(t *testing.T) {
		assert := assert.New(t)

		node.EXPECT().
			GetID().
			Return(peer)

		got, err := c.GetSelfID()

		assert.NoError(err)
		assert.Equal(peer, got)
	})

	t.Run("get balance of", func(t *testing.T) {
		assert := assert.New(t)

		amount := rand.Uint64()

		otherPeer := hash.FakePeer()
		node.EXPECT().
			GetID().
			Return(otherPeer)
		consensus.EXPECT().
			GetBalanceOf(peer).
			Return(amount)

		got, err := c.GetBalanceOf(peer)
		if !assert.NoError(err) {
			return
		}

		expect := &Balance{
			Amount: amount,
		}
		assert.Equal(expect, got)
	})

	t.Run("get balance of self", func(t *testing.T) {
		assert := assert.New(t)

		amount := rand.Uint64()

		node.EXPECT().
			GetID().
			Return(peer)
		consensus.EXPECT().
			GetBalanceOf(peer).
			Return(amount)
		otherPeer := hash.FakePeer()
		interTxn := inter.InternalTransaction{
			Amount:   20,
			Receiver: otherPeer,
		}
		node.EXPECT().
			GetInternalTxns().
			Return([]*inter.InternalTransaction{
				&interTxn,
			})

		got, err := c.GetBalanceOf(peer)
		if !assert.NoError(err) {
			return
		}

		expect := &Balance{
			Amount: amount,
			Pending: []inter.InternalTransaction{
				interTxn,
			},
		}
		assert.Equal(expect, got)
	})

	t.Run("send to", func(t *testing.T) {
		assert := assert.New(t)

		amount := rand.Uint64()

		tx := inter.InternalTransaction{
			Amount:   amount,
			Receiver: peer,
		}

		node.EXPECT().
			GetID().
			Return(peer)
		consensus.EXPECT().
			GetBalanceOf(peer).
			Return(amount)
		node.EXPECT().
			AddInternalTxn(tx)

		err := c.SendTo(tx.Receiver, tx.Amount)
		assert.NoError(err)
	})

	t.Run("send to insufficient", func(t *testing.T) {
		assert := assert.New(t)

		amount := rand.Uint64()

		tx := inter.InternalTransaction{
			Amount:   amount,
			Receiver: peer,
		}

		node.EXPECT().
			GetID().
			Return(peer)
		consensus.EXPECT().
			GetBalanceOf(peer).
			Return(amount - 1)

		err := c.SendTo(tx.Receiver, tx.Amount)
		assert.Error(err)
	})

	t.Run("send to zero amount", func(t *testing.T) {
		assert := assert.New(t)

		tx := inter.InternalTransaction{
			Amount:   0,
			Receiver: peer,
		}

		err := c.SendTo(tx.Receiver, tx.Amount)
		assert.Error(err)
	})
}
