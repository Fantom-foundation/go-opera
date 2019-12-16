package vector

import (
	"testing"

	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/pos"

	"github.com/Fantom-foundation/go-lachesis/kvdb/memorydb"
)

var (
	testASCIIScheme = `
a1.0   b1.0   c1.0   d1.0   e1.0
║      ║      ║      ║      ║
║      ╠──────╫───── d2.0   ║
║      ║      ║      ║      ║
║      b2.1 ──╫──────╣      e2.1
║      ║      ║      ║      ║
║      ╠──────╫───── d3.1   ║
a2.1 ──╣      ║      ║      ║
║      ║      ║      ║      ║
║      b3.2 ──╣      ║      ║
║      ║      ║      ║      ║
║      ╠──────╫───── d4.2   ║
║      ║      ║      ║      ║
║      ╠───── c2.2   ║      e3.2
║      ║      ║      ║      ║
`
)

func BenchmarkIndex_Add(b *testing.B) {
	b.StopTimer()
	ordered := make([]*inter.Event, 0)
	nodes, _, _ := inter.ASCIIschemeForEach(testASCIIScheme, inter.ForEachEvent{
		Process: func(e *inter.Event, name string) {
			ordered = append(ordered, e)
		},
	})
	validatorsBuilder := pos.NewBuilder()
	for _, peer := range nodes {
		validatorsBuilder.Set(peer, 1)
	}
	validators := validatorsBuilder.Build()
	events := make(map[hash.Event]*inter.EventHeaderData)
	getEvent := func(id hash.Event) *inter.EventHeaderData {
		return events[id]
	}
	for _, e := range ordered {
		events[e.Hash()] = &e.EventHeaderData
	}

	vecClock := NewIndex(DefaultIndexConfig(), validators, memorydb.New(), getEvent)

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		vecClock.Reset(validators, memorydb.New(), getEvent)
		b.StartTimer()
		for _, e := range ordered {
			vecClock.Add(&e.EventHeaderData)
			i++
			if i >= b.N {
				break
			}
		}
	}
}
