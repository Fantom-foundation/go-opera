package vector

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/pos"
	"github.com/Fantom-foundation/go-lachesis/src/kvdb/memorydb"
	"github.com/Fantom-foundation/go-lachesis/src/logger"
)

func testMedianTime(t *testing.T, dag string, weights []pos.Stake, claimedTimes map[string]inter.Timestamp, medianTimes map[string]inter.Timestamp, genesis inter.Timestamp) {
	logger.SetTestMode(t)
	assertar := assert.New(t)

	ordered := make([]*inter.Event, 0)
	peers, _, named := inter.ASCIIschemeForEach(dag, inter.ForEachEvent{
		Build: func(e *inter.Event, name string) *inter.Event {
			e.ClaimedTime = claimedTimes[name]
			return e
		},
		Process: func(e *inter.Event, name string) {
			ordered = append(ordered, e)
		},
	})

	members := make(pos.Members, len(peers))
	for i, peer := range peers {
		members.Set(peer, weights[i])
	}

	events := make(map[hash.Event]*inter.EventHeaderData)
	getEvent := func(id hash.Event) *inter.EventHeaderData {
		return events[id]
	}

	vi := NewIndex(members, memorydb.New(), getEvent)

	// push
	for _, e := range ordered {
		events[e.Hash()] = &e.EventHeaderData
		vi.Add(&e.EventHeaderData)
		vi.Flush()
	}

	// check
	for name, e := range named {
		expected, ok := medianTimes[name]
		if !ok {
			continue
		}
		assertar.Equal(expected, vi.MedianTime(e.Hash(), genesis), name)
	}
}

func TestMedianTimeAscii(t *testing.T) {
	dagStr := `
 ║
 nodeA001
 ║
 nodeA012
 ║            ║
 ║            nodeB001
 ║            ║            ║
 ║            ╠═══════════ nodeC001
 ║║           ║            ║            ║
 ║╚══════════─╫─══════════─╫─══════════ nodeD001
║║            ║            ║            ║
╚ nodeA002════╬════════════╬════════════╣
 ║║           ║            ║            ║
 ║╚══════════─╫─══════════─╫─══════════ nodeD002
 ║            ║            ║            ║
 nodeA003════─╫─══════════─╫─═══════════╣
 ║            ║            ║
 ╠════════════nodeB002     ║
 ║                         ║
 ╠═════════════════════════nodeC002
`

	weights := []pos.Stake{3, 4, 2, 1}
	genesisTime := inter.Timestamp(1)
	claimedTimes := map[string]inter.Timestamp{
		"nodeA001": inter.Timestamp(111),
		"nodeB001": inter.Timestamp(112),
		"nodeC001": inter.Timestamp(13),
		"nodeD001": inter.Timestamp(14),
		"nodeA002": inter.Timestamp(120),
		"nodeD002": inter.Timestamp(20),
		"nodeA012": inter.Timestamp(120),
		"nodeA003": inter.Timestamp(131),
		"nodeB002": inter.Timestamp(124),
		"nodeC002": inter.Timestamp(20),
	}
	medianTimes := map[string]inter.Timestamp{
		"nodeA001": genesisTime,
		"nodeB001": genesisTime,
		"nodeC001": inter.Timestamp(13),
		"nodeD001": genesisTime,
		"nodeA002": inter.Timestamp(112),
		"nodeD002": genesisTime,
		"nodeA012": genesisTime,
		"nodeA003": inter.Timestamp(112),
		"nodeB002": inter.Timestamp(124),
		"nodeC002": inter.Timestamp(20),
	}
	t.Run("medianTimeWithForks", func(t *testing.T) {
		testMedianTime(t, dagStr, weights, claimedTimes, medianTimes, genesisTime)
	})
}

func TestMedianTime(t *testing.T) {
	peers := inter.GenNodes(5)
	members := make(pos.Members, len(peers))

	weights := []pos.Stake{5, 4, 3, 2, 1}
	for i, peer := range peers {
		members.Set(peer, weights[i])
	}

	vi := NewIndex(members, memorydb.New(), nil)

	assertar := assert.New(t)
	{ // seq=0
		e := inter.NewEvent().EventHeaderData.Hash()
		// member indexes are sorted by stake amount
		beforeSee := NewHighestBeforeSeq(len(members))
		beforeTime := NewHighestBeforeTime(len(members))

		beforeSee.Set(0, ForkSeq{Seq: 0})
		beforeTime.Set(0, 100)

		beforeSee.Set(1, ForkSeq{Seq: 0})
		beforeTime.Set(1, 100)

		beforeSee.Set(2, ForkSeq{Seq: 1})
		beforeTime.Set(2, 10)

		beforeSee.Set(3, ForkSeq{Seq: 1})
		beforeTime.Set(3, 10)

		beforeSee.Set(4, ForkSeq{Seq: 1})
		beforeTime.Set(4, 10)

		vi.SetHighestBefore(e, beforeSee, beforeTime)
		assertar.Equal(inter.Timestamp(1), vi.MedianTime(e, 1))
	}

	{ // fork seen = true
		e := inter.NewEvent().EventHeaderData.Hash()
		// member indexes are sorted by stake amount
		beforeSee := NewHighestBeforeSeq(len(members))
		beforeTime := NewHighestBeforeTime(len(members))

		beforeSee.Set(0, ForkSeq{Seq: 0, IsForkSeen: true})
		beforeTime.Set(0, 100)

		beforeSee.Set(1, ForkSeq{Seq: 0, IsForkSeen: true})
		beforeTime.Set(1, 100)

		beforeSee.Set(2, ForkSeq{Seq: 1})
		beforeTime.Set(2, 10)

		beforeSee.Set(3, ForkSeq{Seq: 1})
		beforeTime.Set(3, 10)

		beforeSee.Set(4, ForkSeq{Seq: 1})
		beforeTime.Set(4, 10)

		vi.SetHighestBefore(e, beforeSee, beforeTime)
		assertar.Equal(inter.Timestamp(10), vi.MedianTime(e, 1))
	}

	{ // normal
		e := inter.NewEvent().EventHeaderData.Hash()
		// member indexes are sorted by stake amount
		beforeSee := NewHighestBeforeSeq(len(members))
		beforeTime := NewHighestBeforeTime(len(members))

		beforeSee.Set(0, ForkSeq{Seq: 1})
		beforeTime.Set(0, 11)

		beforeSee.Set(1, ForkSeq{Seq: 2})
		beforeTime.Set(1, 12)

		beforeSee.Set(2, ForkSeq{Seq: 2})
		beforeTime.Set(2, 13)

		beforeSee.Set(3, ForkSeq{Seq: 3})
		beforeTime.Set(3, 14)

		beforeSee.Set(4, ForkSeq{Seq: 4})
		beforeTime.Set(4, 15)

		vi.SetHighestBefore(e, beforeSee, beforeTime)
		assertar.Equal(inter.Timestamp(12), vi.MedianTime(e, 1))
	}

}
