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

		consensus.EXPECT().
			GetBalanceOf(peer).
			Return(amount)

		got, err := c.GetBalanceOf(peer)
		if !assert.NoError(err) {
			return
		}

		expect := amount
		assert.Equal(expect, got)
	})

	t.Run("transaction not found", func(t *testing.T) {
		assert := assert.New(t)

		amount := rand.Uint64()
		sender := hash.FakePeer()
		iTx := inter.InternalTransaction{
			Index:    1,
			Amount:   amount,
			Receiver: peer,
			Sender:   sender,
		}

		consensus.EXPECT().
			GetTransaction(iTx.Hash()).
			Return(nil)

		_, err := c.GetTransaction(iTx.Hash())
		assert.Error(err)
	})

	t.Run("transaction", func(t *testing.T) {
		assert := assert.New(t)

		amount := rand.Uint64()
		sender := hash.FakePeer()
		iTx := inter.InternalTransaction{
			Index:    1,
			Amount:   amount,
			Receiver: peer,
			Sender:   sender,
		}

		consensus.EXPECT().
			GetTransaction(iTx.Hash()).
			Return(&iTx)

		got, err := c.GetTransaction(iTx.Hash())
		if !assert.NoError(err) {
			return
		}

		expect := Transaction{
			Amount:   amount,
			Receiver: peer,
			Sender:   sender,
		}

		assertTransactions(assert, &expect, got)
	})

	t.Run("get balance of self", func(t *testing.T) {
		assert := assert.New(t)

		amount := rand.Uint64()

		consensus.EXPECT().
			GetBalanceOf(peer).
			Return(amount)

		got, err := c.GetBalanceOf(peer)
		if !assert.NoError(err) {
			return
		}

		expect := amount
		assert.Equal(expect, got)
	})

	t.Run("send to", func(t *testing.T) {
		assert := assert.New(t)

		amount := rand.Uint64()

		tx := inter.InternalTransaction{
			Amount:   amount,
			Receiver: peer,
		}

		sender := hash.FakePeer()
		node.EXPECT().
			GetID().
			Return(sender)
		consensus.EXPECT().
			GetBalanceOf(sender).
			Return(amount)
		node.EXPECT().
			AddInternalTxn(tx)

		_, err := c.SendTo(tx.Receiver, tx.Amount)
		assert.NoError(err)
	})

	t.Run("send to self", func(t *testing.T) {
		assert := assert.New(t)

		amount := rand.Uint64()

		tx := inter.InternalTransaction{
			Amount:   amount,
			Receiver: peer,
		}

		node.EXPECT().
			GetID().
			Return(peer)

		_, err := c.SendTo(tx.Receiver, tx.Amount)
		assert.Error(err)
	})

	t.Run("send to insufficient", func(t *testing.T) {
		assert := assert.New(t)

		amount := rand.Uint64()

		tx := inter.InternalTransaction{
			Amount:   amount,
			Receiver: peer,
		}

		sender := hash.FakePeer()
		node.EXPECT().
			GetID().
			Return(sender)
		consensus.EXPECT().
			GetBalanceOf(sender).
			Return(amount - 1)

		_, err := c.SendTo(tx.Receiver, tx.Amount)
		assert.Error(err)
	})

	t.Run("send to zero amount", func(t *testing.T) {
		assert := assert.New(t)

		tx := inter.InternalTransaction{
			Amount:   0,
			Receiver: peer,
		}

		_, err := c.SendTo(tx.Receiver, tx.Amount)
		assert.Error(err)
	})
}

func assertTransactions(assert *assert.Assertions, expect, got *Transaction) {
	assert.Equal(expect.Amount, got.Amount)
	assert.Equal(expect.Receiver.Hex(), got.Receiver.Hex())
	assert.Equal(expect.Sender.Hex(), got.Sender.Hex())
	assert.Equal(expect.Confirmed, got.Confirmed)
}
