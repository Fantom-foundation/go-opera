package proxy

import (
	"math/rand"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/pos"
	"github.com/Fantom-foundation/go-lachesis/src/logger"
	"github.com/Fantom-foundation/go-lachesis/src/network"
)

func TestGrpcCtrlCalls(t *testing.T) {
	logger.SetTestMode(t)

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

	s, addr, err := NewGrpcCtrlProxy("127.0.0.1:", node, consensus, nil)
	if !assert.NoError(t, err) {
		return
	}
	defer s.Close()

	client, err := NewGrpcNodeProxy(addr)
	if !assert.NoError(t, err) {
		return
	}
	defer client.Close()

	peer := hash.FakePeer()

	t.Run("get self id", func(t *testing.T) {
		assertar := assert.New(t)

		got, err := client.GetSelfID()

		assertar.NoError(err)
		assertar.Equal(id, got)
	})

	t.Run("get balance of", func(t *testing.T) {
		assertar := assert.New(t)

		expect := pos.Stake(rand.Uint64())

		consensus.EXPECT().
			StakeOf(peer).
			Return(expect)

		got, err := client.StakeOf(peer)
		if !assertar.NoError(err) {
			return
		}

		assertar.Equal(expect, got)
	})

	t.Run("transaction not found", func(t *testing.T) {
		assertar := assert.New(t)

		h := hash.FakeTransaction()

		node.EXPECT().
			GetInternalTxn(h).
			Return(nil, nil)

		_, _, _, err := client.GetTxnInfo(h)
		assertar.Error(err)
	})

	t.Run("transaction", func(t *testing.T) {
		assertar := assert.New(t)

		h := hash.FakeTransaction()
		txn0 := &inter.InternalTransaction{
			Nonce:    1,
			Amount:   pos.Stake(rand.Uint64()),
			Receiver: peer,
		}
		event0 := inter.NewEvent()
		event0.Seq = 1
		event0.Creator = hash.FakePeer()
		event0.Parents = hash.Events{hash.ZeroEvent}

		node.EXPECT().
			GetInternalTxn(h).
			Return(txn0, event0)

		consensus.EXPECT().
			GetEventBlock(event0.Hash()).
			Return(nil)

		txn1, event1, _, err := client.GetTxnInfo(h)
		if !assertar.NoError(err) {
			return
		}

		assertar.EqualValues(txn0, txn1)
		assertar.EqualValues(event0.Hash(), event1.Hash())
	})

	t.Run("get balance of self", func(t *testing.T) {
		assertar := assert.New(t)

		expect := pos.Stake(rand.Uint64())

		consensus.EXPECT().
			StakeOf(peer).
			Return(expect)

		got, err := client.StakeOf(peer)
		if !assertar.NoError(err) {
			return
		}

		assertar.Equal(expect, got)
	})

	t.Run("send to", func(t *testing.T) {
		assertar := assert.New(t)

		amount := pos.Stake(rand.Uint64())
		tx := inter.InternalTransaction{
			Nonce:    1,
			Amount:   amount,
			Receiver: peer,
		}

		node.EXPECT().
			AddInternalTxn(tx)

		_, err := client.SendTo(tx.Receiver, tx.Nonce, tx.Amount, tx.UntilBlock)
		assertar.NoError(err)
	})
}
