package seeing

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/ordering"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/internal"
)

func TestStronglySeen(t *testing.T) {
	testStronglySeen(t, `
a0_0  b0_0  c0_0  d0_0
║     ║     ║     ║
a1_1══╬═════╣     ║
║     ║     ║     ║
║╚════b1_1══╣     ║
║     ║     ║     ║
║     ║╚════c1_1══╣
║     ║     ║     ║
║     ║╚═══─╫╩════d1_1
║     ║     ║     ║
a2_2══╬═════╬═════╣
║     ║     ║     ║
`)
}

func testStronglySeen(t *testing.T, dag string) {
	assertar := assert.New(t)

	peers, named := inter.GenEventsByNode(4, 10, 3)
	//peers, _, named := inter.ASCIIschemeToDAG(dag) // NOTE: stub

	mm := make(internal.Members, len(peers))
	for _, peer := range peers {
		mm.Add(peer, inter.Stake(1))
	}

	ss := New(mm)

	processed := make(map[hash.Event]*inter.Event)
	orderThenProcess := ordering.EventBuffer(ordering.Callback{

		Process: func(e *inter.Event) {
			processed[e.Hash()] = e
			ss.Add(e)
		},

		Drop: func(e *inter.Event, err error) {
			t.Fatal(e, err)
		},

		Exists: func(h hash.Event) *inter.Event {
			return processed[h]
		},
	})

	for _, ee := range named {
		for _, e := range ee {
			orderThenProcess(e)
		}
	}

	ss.Reset(mm)

	assertar.NotNil(ss) // NOTE: stub
}
