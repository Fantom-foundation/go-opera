package proxy

import (
	"math/rand"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/logger"
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
	id := hash.FakePeer()
	node.EXPECT().
		GetID().
		Return(id).
		AnyTimes()

	consensus := NewMockConsensus(ctrl)

	s, addr, err := NewGrpcCtrlProxy("127.0.0.1:", node, consensus, nil, nil)
	if !assert.NoError(t, err) {
		return
	}
	defer s.Close()

	client, err := NewGrpcNodeProxy(addr, nil)
	if !assert.NoError(t, err) {
		return
	}
	defer client.Close()

	peer := hash.FakePeer()

	t.Run("get self id", func(t *testing.T) {
		assert := assert.New(t)

		got, err := client.GetSelfID()

		assert.NoError(err)
		assert.Equal(id, got)
	})

	t.Run("get balance of", func(t *testing.T) {
		assert := assert.New(t)

		expect := rand.Uint64()

		consensus.EXPECT().
			StakeOf(peer).
			Return(expect)

		got, err := client.StakeOf(peer)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, got)
	})

	t.Run("transaction not found", func(t *testing.T) {
		assert := assert.New(t)

		h := hash.FakeTransaction()

		consensus.EXPECT().
			GetTransaction(h).
			Return(nil)

		_, err := client.GetTransaction(h)
		assert.Error(err)
	})

	t.Run("transaction", func(t *testing.T) {
		assert := assert.New(t)

		h := hash.FakeTransaction()
		expect := &inter.InternalTransaction{
			Index:    1,
			Amount:   rand.Uint64(),
			Receiver: peer,
		}

		consensus.EXPECT().
			GetTransaction(h).
			Return(expect)

		got, err := client.GetTransaction(h)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, got)
	})

	t.Run("get balance of self", func(t *testing.T) {
		assert := assert.New(t)

		expect := rand.Uint64()

		consensus.EXPECT().
			StakeOf(peer).
			Return(expect)

		got, err := client.StakeOf(peer)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, got)
	})

	t.Run("send to", func(t *testing.T) {
		assert := assert.New(t)

		amount := rand.Uint64()
		tx := inter.InternalTransaction{
			Index:    1,
			Amount:   amount,
			Receiver: peer,
		}

		node.EXPECT().
			AddInternalTxn(tx)

		_, err := client.SendTo(tx.Receiver, tx.Index, tx.Amount, tx.UntilBlock)
		assert.NoError(err)
	})

	t.Run("set log level", func(t *testing.T) {
		assert := assert.New(t)

		l := "info"
		err := client.SetLogLevel(l)
		assert.NoError(err)
		assert.Equal(logger.Get().GetLevel(), logger.GetLevel(l))
	})
}
