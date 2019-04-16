package posnode

import (
	"sort"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
)

func TestEmit(t *testing.T) {
	// node 1
	store1 := NewMemStore()
	node1 := NewForTests("emitter1", store1, nil)

	// node 2
	store2 := NewMemStore()
	node2 := NewForTests("emitter2", store2, nil)

	events := make([]*inter.Event, 3)

	t.Run("no parent events", func(t *testing.T) {
		assert := assert.New(t)

		tx := []byte("12345")
		node1.AddExternalTxn(tx)
		events[0] = node1.EmitEvent()

		assert.Equal(
			uint64(1),
			events[0].Index)
		assert.Equal(
			inter.Timestamp(1),
			events[0].LamportTime)
		assert.Equal(
			hash.NewEvents(hash.ZeroEvent),
			events[0].Parents)
		assert.Equal(
			[][]byte{tx},
			events[0].ExternalTransactions)
	})

	t.Run("zero event", func(t *testing.T) {
		assert := assert.New(t)

		// node2 got event0
		store2.BootstrapPeers(node1.AsPeer())
		node2.initPeers()
		node2.saveNewEvent(events[0])

		events[1] = node2.EmitEvent()

		assert.Equal(
			uint64(1),
			events[1].Index)
		assert.Equal(
			inter.Timestamp(2),
			events[1].LamportTime)
		assert.Equal(
			hash.NewEvents(hash.ZeroEvent, events[0].Hash()),
			events[1].Parents)
	})

	t.Run("has self parent", func(t *testing.T) {
		assert := assert.New(t)

		// node1 got event1
		store1.BootstrapPeers(node2.AsPeer())
		node1.initPeers()
		node1.saveNewEvent(events[1])

		events[2] = node1.EmitEvent()

		assert.Equal(
			uint64(2),
			events[2].Index)
		assert.Equal(
			inter.Timestamp(3),
			events[2].LamportTime)
		assert.Equal(
			hash.NewEvents(events[0].Hash(), events[1].Hash()),
			events[2].Parents)
	})
}

func Test_emitterEvaluation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	consensus := NewMockConsensus(ctrl)
	consensus.EXPECT().PushEvent(gomock.Any()).AnyTimes()

	store := NewMemStore()
	node := NewForTests("server.fake", store, consensus)

	peer1 := Peer{
		ID:     hash.HexToPeer("1"),
		Host:   "host1",
		PubKey: &common.PublicKey{},
	}
	peer2 := Peer{
		ID:     hash.HexToPeer("2"),
		Host:   "host2",
		PubKey: &common.PublicKey{},
	}
	peer3 := Peer{
		ID:     hash.HexToPeer("3"),
		Host:   "host3",
		PubKey: &common.PublicKey{},
	}

	store.BootstrapPeers(&peer1, &peer2, &peer3)
	node.initPeers()
	node.peers.peers[peer1.ID] = &peerAttr{}
	node.peers.peers[peer2.ID] = &peerAttr{}
	node.peers.peers[peer3.ID] = &peerAttr{}

	t.Run("first round", func(t *testing.T) {
		assert := assert.New(t)

		consensus.EXPECT().GetStakeOf(peer1.ID).Return(float64(1)).AnyTimes()
		consensus.EXPECT().GetStakeOf(peer2.ID).Return(float64(2)).AnyTimes()
		consensus.EXPECT().GetStakeOf(peer3.ID).Return(float64(3)).AnyTimes()

		p1ev1 := inter.Event{
			Index:       1,
			Creator:     peer1.ID,
			Parents:     hash.NewEvents(hash.ZeroEvent),
			LamportTime: 1,
		}

		p2ev1 := inter.Event{
			Index:       1,
			Creator:     peer2.ID,
			Parents:     hash.NewEvents(hash.ZeroEvent),
			LamportTime: 1,
		}

		p3ev1 := inter.Event{
			Index:       1,
			Creator:     peer3.ID,
			Parents:     hash.NewEvents(hash.ZeroEvent),
			LamportTime: 1,
		}

		node.saveNewEvent(&p1ev1)
		node.saveNewEvent(&p2ev1)
		node.saveNewEvent(&p3ev1)

		e := node.emitterEvaluation(node.Snapshot())
		sort.Sort(e)

		assert.Equal(e.peers[0], peer3.ID)
		assert.Equal(e.peers[1], peer2.ID)
		assert.Equal(e.peers[2], peer1.ID)
	})

	t.Run("second round", func(t *testing.T) {
		assert := assert.New(t)

		consensus.EXPECT().GetStakeOf(peer1.ID).Return(float64(1)).AnyTimes()
		consensus.EXPECT().GetStakeOf(peer2.ID).Return(float64(2)).AnyTimes()
		consensus.EXPECT().GetStakeOf(peer3.ID).Return(float64(3)).AnyTimes()

		p3ev1 := store.LastEvent(peer3.ID)
		p2ev1 := store.LastEvent(peer2.ID)

		ev := inter.Event{
			Index:       1,
			Creator:     node.ID,
			Parents:     hash.NewEvents(hash.ZeroEvent, p3ev1.Hash(), p2ev1.Hash()),
			LamportTime: 2,
		}

		store.SetEvent(&ev)

		p1ev2 := inter.Event{
			Index:       2,
			Creator:     peer1.ID,
			Parents:     hash.NewEvents(hash.ZeroEvent, p3ev1.Hash(), p2ev1.Hash()),
			LamportTime: 2,
		}

		p2ev2 := inter.Event{
			Index:       2,
			Creator:     peer2.ID,
			Parents:     hash.NewEvents(hash.ZeroEvent, p3ev1.Hash(), p2ev1.Hash()),
			LamportTime: 2,
		}

		p3ev2 := inter.Event{
			Index:       2,
			Creator:     peer3.ID,
			Parents:     hash.NewEvents(hash.ZeroEvent, p3ev1.Hash(), p2ev1.Hash()),
			LamportTime: 2,
		}

		node.saveNewEvent(&p1ev2)
		node.saveNewEvent(&p2ev2)
		node.saveNewEvent(&p3ev2)

		e := node.emitterEvaluation(node.Snapshot())
		sort.Sort(e)

		assert.Equal(e.peers[0], peer3.ID)
		assert.Equal(e.peers[1], peer2.ID)
		assert.Equal(e.peers[2], peer1.ID)
	})

}
