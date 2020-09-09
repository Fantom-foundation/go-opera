package vecmt

import (
	"testing"

	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/vecfc"
	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/dag"
	"github.com/Fantom-foundation/lachesis-base/inter/dag/tdag"
	"github.com/Fantom-foundation/lachesis-base/inter/pos"
	"github.com/Fantom-foundation/lachesis-base/kvdb/memorydb"

	"github.com/Fantom-foundation/go-opera/inter"
)

func TestMedianTimeOnIndex(t *testing.T) {
	nodes := tdag.GenNodes(5)
	weights := []pos.Weight{5, 4, 3, 2, 1}
	validators := pos.ArrayToValidators(nodes, weights)

	vi := NewIndex(func(err error) { panic(err) }, LiteConfig())
	vi.Reset(validators, memorydb.New(), nil)

	assertar := assert.New(t)
	{ // seq=0
		e := hash.ZeroEvent
		// validator indexes are sorted by weight amount
		before := NewHighestBefore(idx.Validator(validators.Len()))

		before.VSeq.Set(0, vecfc.BranchSeq{Seq: 0})
		before.VTime.Set(0, 100)

		before.VSeq.Set(1, vecfc.BranchSeq{Seq: 0})
		before.VTime.Set(1, 100)

		before.VSeq.Set(2, vecfc.BranchSeq{Seq: 1})
		before.VTime.Set(2, 10)

		before.VSeq.Set(3, vecfc.BranchSeq{Seq: 1})
		before.VTime.Set(3, 10)

		before.VSeq.Set(4, vecfc.BranchSeq{Seq: 1})
		before.VTime.Set(4, 10)

		vi.SetHighestBefore(e, before)
		assertar.Equal(inter.Timestamp(1), vi.MedianTime(e, 1))
	}

	{ // fork seen = true
		e := hash.ZeroEvent
		// validator indexes are sorted by weight amount
		before := NewHighestBefore(idx.Validator(validators.Len()))

		before.SetForkDetected(0)
		before.VTime.Set(0, 100)

		before.SetForkDetected(1)
		before.VTime.Set(1, 100)

		before.VSeq.Set(2, vecfc.BranchSeq{Seq: 1})
		before.VTime.Set(2, 10)

		before.VSeq.Set(3, vecfc.BranchSeq{Seq: 1})
		before.VTime.Set(3, 10)

		before.VSeq.Set(4, vecfc.BranchSeq{Seq: 1})
		before.VTime.Set(4, 10)

		vi.SetHighestBefore(e, before)
		assertar.Equal(inter.Timestamp(10), vi.MedianTime(e, 1))
	}

	{ // normal
		e := hash.ZeroEvent
		// validator indexes are sorted by weight amount
		before := NewHighestBefore(idx.Validator(validators.Len()))

		before.VSeq.Set(0, vecfc.BranchSeq{Seq: 1})
		before.VTime.Set(0, 11)

		before.VSeq.Set(1, vecfc.BranchSeq{Seq: 2})
		before.VTime.Set(1, 12)

		before.VSeq.Set(2, vecfc.BranchSeq{Seq: 2})
		before.VTime.Set(2, 13)

		before.VSeq.Set(3, vecfc.BranchSeq{Seq: 3})
		before.VTime.Set(3, 14)

		before.VSeq.Set(4, vecfc.BranchSeq{Seq: 4})
		before.VTime.Set(4, 15)

		vi.SetHighestBefore(e, before)
		assertar.Equal(inter.Timestamp(12), vi.MedianTime(e, 1))
	}

}

func TestMedianTimeOnDAG(t *testing.T) {
	dagAscii := `
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
 ║            ║            ║
 ╠════════════╫═══════════ nodeC002
`

	weights := []pos.Weight{3, 4, 2, 1}
	genesisTime := inter.Timestamp(1)
	creationTimes := map[string]inter.Timestamp{
		"nodeA001": inter.Timestamp(111),
		"nodeB001": inter.Timestamp(112),
		"nodeC001": inter.Timestamp(13),
		"nodeD001": inter.Timestamp(14),
		"nodeA002": inter.Timestamp(120),
		"nodeD002": inter.Timestamp(20),
		"nodeA012": inter.Timestamp(120),
		"nodeA003": inter.Timestamp(20),
		"nodeB002": inter.Timestamp(20),
		"nodeC002": inter.Timestamp(35),
	}
	medianTimes := map[string]inter.Timestamp{
		"nodeA001": genesisTime,
		"nodeB001": genesisTime,
		"nodeC001": inter.Timestamp(13),
		"nodeD001": genesisTime,
		"nodeA002": inter.Timestamp(112),
		"nodeD002": genesisTime,
		"nodeA012": genesisTime,
		"nodeA003": inter.Timestamp(20),
		"nodeB002": inter.Timestamp(20),
		"nodeC002": inter.Timestamp(35),
	}
	t.Run("testMedianTimeOnDAG", func(t *testing.T) {
		testMedianTime(t, dagAscii, weights, creationTimes, medianTimes, genesisTime)
	})
}

func testMedianTime(t *testing.T, dagAscii string, weights []pos.Weight, creationTimes map[string]inter.Timestamp, medianTimes map[string]inter.Timestamp, genesis inter.Timestamp) {
	assertar := assert.New(t)

	var ordered dag.Events
	nodes, _, named := tdag.ASCIIschemeForEach(dagAscii, tdag.ForEachEvent{
		Process: func(e dag.Event, name string) {
			ordered = append(ordered, &eventWithCreationTime{e, creationTimes[name]})
		},
	})

	validators := pos.ArrayToValidators(nodes, weights)

	events := make(map[hash.Event]dag.Event)
	getEvent := func(id hash.Event) dag.Event {
		return events[id]
	}

	vi := NewIndex(func(err error) { panic(err) }, LiteConfig())
	vi.Reset(validators, memorydb.New(), getEvent)

	// push
	for _, e := range ordered {
		events[e.ID()] = e
		assertar.NoError(vi.Add(e))
		vi.Flush()
	}

	// check
	for name, e := range named {
		expected, ok := medianTimes[name]
		if !ok {
			continue
		}
		assertar.Equal(expected, vi.MedianTime(e.ID(), genesis), name)
	}
}
