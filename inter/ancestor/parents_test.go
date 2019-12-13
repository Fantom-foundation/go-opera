package ancestor

import (
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/inter/pos"
	"github.com/Fantom-foundation/go-lachesis/kvdb/memorydb"
	"github.com/Fantom-foundation/go-lachesis/logger"
	"github.com/Fantom-foundation/go-lachesis/utils"
	"github.com/Fantom-foundation/go-lachesis/vector"
)

func TestCasualityStrategy(t *testing.T) {
	testSpecialNamedParents(t, `
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
`, map[int]map[string]string{
		0: {
			"nodeA": "[a1.0, c1.0, d2.0, e1.0]",
			"nodeB": "[b1.0, a1.0, c1.0, d2.0, e1.0]",
			"nodeC": "[c1.0, a1.0, d2.0, e1.0]",
			"nodeD": "[d2.0, a1.0, c1.0, e1.0]",
			"nodeE": "[e1.0, a1.0, c1.0, d2.0]",
		},
		1: {
			"nodeA": "[a2.1, c1.0, d3.1, e2.1]",
			"nodeB": "[b2.1, a2.1, c1.0, d3.1, e2.1]",
			"nodeC": "[c1.0, a2.1, d3.1, e2.1]",
			"nodeD": "[d3.1, a2.1, c1.0, e2.1]",
			"nodeE": "[e2.1, a2.1, c1.0, d3.1]",
		},
		2: {
			"nodeA": "[a2.1, c2.2, d4.2, e3.2]",
			"nodeB": "[b3.2, a2.1, c2.2, d4.2, e3.2]",
			"nodeC": "[c2.2, a2.1, d4.2, e3.2]",
			"nodeD": "[d4.2, a2.1, c2.2, e3.2]",
			"nodeE": "[e3.2, a2.1, c2.2, d4.2]",
		},
	})
}

// testSpecialNamedParents is a general test of parent selection.
// Event name means:
// - unique event name;
// - "." - separator;
// - stage - makes ;
func testSpecialNamedParents(t *testing.T, asciiScheme string, exp map[int]map[string]string) {
	logger.SetTestMode(t)
	assertar := assert.New(t)

	// decode is a event name parser
	decode := func(name string) (stage int) {
		n, err := strconv.ParseUint(strings.Split(name, ".")[1], 10, 32)
		if err != nil {
			panic(err.Error() + ". Name event " + name + " properly: <UniqueName>.<StageN>")
		}
		stage = int(n)
		return
	}

	ordered := make([]*inter.Event, 0)
	nodes, _, _ := inter.ASCIIschemeForEach(asciiScheme, inter.ForEachEvent{
		Process: func(e *inter.Event, name string) {
			ordered = append(ordered, e)
		},
	})

	validators := pos.EqualStakeValidators(nodes, 1)

	events := make(map[hash.Event]*inter.EventHeaderData)
	getEvent := func(id hash.Event) *inter.EventHeaderData {
		return events[id]
	}

	vecClock := vector.NewIndex(vector.DefaultIndexConfig(), validators, memorydb.New(), getEvent)

	// build vector index
	for _, e := range ordered {
		events[e.Hash()] = &e.EventHeaderData
		vecClock.Add(&e.EventHeaderData)
	}

	// divide events by stage
	var stages [][]*inter.Event
	for _, e := range ordered {
		name := string(e.Extra)
		stage := decode(name)
		for i := len(stages); i <= stage; i++ {
			stages = append(stages, nil)
		}
		stages[stage] = append(stages[stage], e)
	}

	heads := hash.EventsSet{}
	tips := map[idx.StakerID]*hash.Event{}
	// check
	for stage, ee := range stages {
		t.Logf("Stage %d:", stage)

		// build heads/tips
		for _, e := range ee {
			for _, p := range e.Parents {
				if heads.Contains(p) {
					heads.Erase(p)
				}
			}
			heads.Add(e.Hash())
			id := e.Hash()
			tips[e.Creator] = &id
		}

		for _, node := range nodes {
			selfParent := tips[node]

			strategy := NewCasualityStrategy(vecClock, validators)

			selfParentResult, parents := FindBestParents(5, heads.Slice(), selfParent, strategy)

			if selfParent != nil {
				assertar.Equal(parents[0], *selfParent)
				assertar.Equal(*selfParentResult, *selfParent)
			} else {
				assertar.Nil(selfParentResult)
			}
			//t.Logf("\"%s\": \"%s\",", node.String(), parentsToString(parents))
			if !assertar.Equal(
				exp[stage][utils.NameOf(node)],
				parentsToString(parents),
				"stage %d, %s", stage, utils.NameOf(node),
			) {
				return
			}
		}
	}

	assertar.NoError(nil)
}

func parentsToString(pp hash.Events) string {
	if len(pp) < 3 {
		return pp.String()
	}

	res := make(hash.Events, len(pp))
	copy(res, pp)

	sort.Slice(res[1:], func(i, j int) bool {
		return res[i+1].String() < res[j+1].String()
	})

	return res.String()
}
