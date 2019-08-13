package posposet

import (
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/logger"
)

func TestSeeingStrategy(t *testing.T) {
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

	nodes, _, names := ASCIIschemeToDAG(asciiScheme, nil, nil)

	// divide events by stage
	var stages [][]*inter.Event
	for name, e := range names {
		stage := decode(name)
		for i := len(stages); i <= stage; i++ {
			stages = append(stages, nil)
		}
		stages[stage] = append(stages[stage], e)
	}

	p, _, input := FakePoset(nodes)

	for stage, ee := range stages {
		t.Logf("Stage %d:", stage)
		for _, e := range ee {
			input.SetEvent(e)
			p.PushEventSync(e.Hash())
		}

		for _, node := range nodes {
			strategy := p.NewSeeingStrategy()
			parents := p.FindBestParents(node, 5, strategy)
			//t.Logf("\"%s\": \"%s\",", node.String(), parentsToString(parents))
			//continue
			if !assertar.Equal(
				exp[stage][node.String()],
				parentsToString(parents),
				"stage %d, %s", stage, node.String(),
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
