package posnode

import (
	"bytes"
	"testing"

	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/stretchr/testify/assert"
)

func TestEmit(t *testing.T) {
	t.Run("no transactions", func(t *testing.T) {
		store := NewMemStore()
		n := NewForTests("server.fake", store, nil)

		n.CreateEvent()
	})

	t.Run("not enough events", func(t *testing.T) {
		assert := assert.New(t)

		store := NewMemStore()
		n := NewForTests("server.fake", store, nil)

		var buf bytes.Buffer
		n.logger.log.Logger.Out = &buf
		n.emitter.transactions = [][]byte{
			[]byte{},
		}
		n.CreateEvent()

		assert.Contains(buf.String(), "enough events")
	})

	t.Run("zero event", func(t *testing.T) {
		assert := assert.New(t)

		// node 1
		store1 := NewMemStore()
		node1 := NewForTests("node1", store1, nil)
		node1.StartServiceForTests()
		defer node1.StopService()

		// node 2
		store2 := NewMemStore()
		node2 := NewForTests("node2", store2, nil)
		node2.conf.EventParentsCount = 2

		// connect node2 to node1
		store2.BootstrapPeers(&Peer{
			ID:     node1.ID,
			PubKey: node1.pub,
			Host:   node1.host,
		})
		node2.initPeers()

		// create node 1 event
		creater1 := node1.ToPeer()
		event := inter.Event{
			Index:                1,
			Creator:              creater1,
			LamportTime:          1,
			ExternalTransactions: make([][]byte, 0),
		}

		sign, _ := sign(node1.key, event.Hash().Bytes())
		event.Sign = sign
		node2.SaveNewEvent(&event)
		node2.emitter.transactions = [][]byte{
			[]byte{},
		}

		node2.CreateEvent()

		creater2 := node2.ToPeer()
		index := store2.GetPeerHeight(node2.ID)
		eventHash := store2.GetEventHash(creater2, index)
		got := store2.GetEvent(*eventHash)

		assert.Equal(int(got.LamportTime), 2)
		assert.Equal(int(got.Index), 1)
		assert.Equal(len(got.Parents), 2)
	})

	t.Run("has self parent", func(t *testing.T) {
		assert := assert.New(t)

		// node 1
		store1 := NewMemStore()
		node1 := NewForTests("node1", store1, nil)
		node1.StartServiceForTests()
		defer node1.StopService()

		// node 2
		store2 := NewMemStore()
		node2 := NewForTests("node2", store2, nil)
		node2.conf.EventParentsCount = 2

		// connect node2 to node1
		store2.BootstrapPeers(&Peer{
			ID:     node1.ID,
			PubKey: node1.pub,
			Host:   node1.host,
		})
		node2.initPeers()

		// create node 1 event
		creater1 := node1.ToPeer()
		event1 := inter.Event{
			Index:                1,
			Creator:              creater1,
			LamportTime:          1,
			ExternalTransactions: make([][]byte, 0),
		}

		s1, _ := sign(node1.key, event1.Hash().Bytes())
		event1.Sign = s1
		node2.SaveNewEvent(&event1)

		// create node 2 event
		creater2 := node2.ToPeer()
		event2 := inter.Event{
			Index:                1,
			Creator:              creater2,
			LamportTime:          1,
			ExternalTransactions: make([][]byte, 0),
		}

		s2, _ := sign(node2.key, event2.Hash().Bytes())
		event2.Sign = s2
		node2.SaveNewEvent(&event2)
		node2.emitter.transactions = [][]byte{
			[]byte{},
		}

		node2.CreateEvent()

		creater := node2.ToPeer()
		index := store2.GetPeerHeight(node2.ID)
		eventHash := store2.GetEventHash(creater, index)
		got := store2.GetEvent(*eventHash)

		assert.Equal(int(got.LamportTime), 2)
		assert.Equal(int(got.Index), 2)
		assert.Equal(len(got.Parents), 2)
	})
}
