package posnode

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/logger"
)

func TestAddInternalTxn(t *testing.T) {
	logger.SetTestMode(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	consensus := NewMockConsensus(ctrl)
	consensus.EXPECT().
		StakeOf(gomock.Any()).
		Return(uint64(2000)).
		AnyTimes()

	node := NewForTests("fake", NewMemStore(), consensus)
	peer := hash.FakePeer()

	t.Run("very 1st add", func(t *testing.T) {
		assertar := assert.New(t)

		tx := inter.InternalTransaction{
			Index:    1,
			Amount:   1000,
			Receiver: peer,
		}

		_, err := node.AddInternalTxn(tx)
		if !assertar.NoError(err) {
			return
		}
		// TODO: check when implemented
		//assertar.Equal(expect, h.Hex())
	})

	t.Run("very 2nd add", func(t *testing.T) {
		assertar := assert.New(t)

		tx := inter.InternalTransaction{
			Index:    2,
			Amount:   1000,
			Receiver: peer,
		}

		_, err := node.AddInternalTxn(tx)
		if !assertar.NoError(err) {
			return
		}
		// TODO: check when implemented
		//assert.Equal(expect, h.Hex())
	})
}

func TestEmit(t *testing.T) {
	logger.SetTestMode(t)

	// node 1
	store1 := NewMemStore()
	node1 := NewForTests("emitter1", store1, nil)
	node1.initParents()
	defer node1.Stop()

	// node 2
	store2 := NewMemStore()
	node2 := NewForTests("emitter2", store2, nil)
	node2.initParents()
	defer node2.Stop()

	events := make([]*inter.Event, 4)

	t.Run("very 1st event", func(t *testing.T) {
		assertar := assert.New(t)
		// node1 has no candidates to parent
		tx := []byte("12345")
		node1.AddExternalTxn(tx)

		events[0] = node1.EmitEvent()

		assertar.Equal(
			uint64(1),
			events[0].Index)
		assertar.Equal(
			inter.Timestamp(1),
			events[0].LamportTime)
		assertar.Equal(
			hash.NewEvents(hash.ZeroEvent),
			events[0].Parents)
		assertar.Equal(
			[][]byte{tx},
			events[0].ExternalTransactions)
	})

	t.Run("1st event", func(t *testing.T) {
		assertar := assert.New(t)
		// node2 got event0
		node2.onNewEvent(events[0])

		events[1] = node2.EmitEvent()

		assertar.Equal(
			uint64(1),
			events[1].Index)
		assertar.Equal(
			inter.Timestamp(2),
			events[1].LamportTime)
		assertar.Equal(
			hash.NewEvents(hash.ZeroEvent, events[0].Hash()),
			events[1].Parents)
	})

	t.Run("2nd event", func(t *testing.T) {
		assertar := assert.New(t)
		// node1 got event1
		node1.onNewEvent(events[1])

		events[2] = node1.EmitEvent()

		assertar.Equal(
			uint64(2),
			events[2].Index)
		assertar.Equal(
			inter.Timestamp(3),
			events[2].LamportTime)
		assertar.Equal(
			hash.NewEvents(events[0].Hash(), events[1].Hash()),
			events[2].Parents)
	})

	t.Run("3rd event", func(t *testing.T) {
		assertar := assert.New(t)
		// node1 has no new parents
		// TODO: what about skip event creation?

		events[3] = node1.EmitEvent()

		assertar.Equal(
			uint64(3),
			events[3].Index)
		assertar.Equal(
			inter.Timestamp(4),
			events[3].LamportTime)
		assertar.Equal(
			hash.NewEvents(events[2].Hash()),
			events[3].Parents)
	})
}
