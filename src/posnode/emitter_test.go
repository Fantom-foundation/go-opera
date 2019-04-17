package posnode

import (
	"sort"
	"testing"
	"time"

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

	consensus.EXPECT().GetStakeOf(peer1.ID).Return(float64(1)).AnyTimes()
	consensus.EXPECT().GetStakeOf(peer2.ID).Return(float64(2)).AnyTimes()
	consensus.EXPECT().GetStakeOf(peer3.ID).Return(float64(3)).AnyTimes()

	store.BootstrapPeers(&peer1, &peer2, &peer3)
	node.initPeers()
	node.peers.ids[peer1.ID] = &peerAttr{}
	node.peers.ids[peer2.ID] = &peerAttr{}
	node.peers.ids[peer3.ID] = &peerAttr{}

	t.Run("last used", func(t *testing.T) {
		assert := assert.New(t)

		node.peers.ids[peer1.ID].LastUsed = time.Now().Add(2 * time.Hour)
		node.peers.ids[peer2.ID].LastUsed = time.Now().Add(time.Hour)
		node.peers.ids[peer3.ID].LastUsed = time.Now()

		e := node.emitterEvaluation(node.Snapshot())
		sort.Sort(e)

		assert.Equal(e.peers[0], peer3.ID)
		assert.Equal(e.peers[1], peer2.ID)
		assert.Equal(e.peers[2], peer1.ID)
	})

	t.Run("last event", func(t *testing.T) {
		assert := assert.New(t)

		node.peers.ids[peer3.ID].LastEvent = time.Now().Add(2 * time.Hour)
		node.peers.ids[peer2.ID].LastEvent = time.Now().Add(time.Hour)
		node.peers.ids[peer1.ID].LastEvent = time.Now()

		e := node.emitterEvaluation(node.Snapshot())
		sort.Sort(e)

		assert.Equal(e.peers[0], peer3.ID)
		assert.Equal(e.peers[1], peer2.ID)
		assert.Equal(e.peers[2], peer1.ID)
	})

	t.Run("balance", func(t *testing.T) {
		assert := assert.New(t)

		e := node.emitterEvaluation(node.Snapshot())
		sort.Sort(e)

		assert.Equal(e.peers[0], peer3.ID)
		assert.Equal(e.peers[1], peer2.ID)
		assert.Equal(e.peers[2], peer1.ID)
	})
}
