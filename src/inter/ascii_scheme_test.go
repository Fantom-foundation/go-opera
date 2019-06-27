package inter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/src/utils"
)

func TestASCIIschemeToDAG(t *testing.T) {
	nodes, _, named := ASCIIschemeToDAG(`
a00 b00   c00 d00
║   ║     ║   ║
a01 ║     ║   ║
║   ╠  ─  c01 ║
a02 ╣     ║   ║
║   ║     ║   ║
╠ ─ ╫ ─ ─ c02 ║
║   b01  ╝║   ║
║   ╠ ─ ─ ╫ ─ d01
║   ║     ║   ║
║   ║     ║   ║
╠ ═ b02═══╬   ╣
║   ║     ║  ║║
a03 ╣     ╠ ─ d02
║║  ║     ║  ║║
║║  ║     ║  ║╠ ─ e00
║║  ║     ║   ║   ║
a04 ╫ ─ ─ ╬  ╝║   ║
║║  ║     ║   ║   ║
║╚═─╫╩  ─ c03 ╣   ║
║   ║     ║   ║   ║
║   ║     ║  3║   ║
║   b03 ─ ╫  ╝║   ║
║   ║     ║   ║   ║
`)
	expected := map[string][]string{
		"a00": {""},
		"a01": {"a00"},
		"a02": {"a01", "b00"},
		"a03": {"a02", "b02"},
		"a04": {"a03", "c02", "d01"},
		"b00": {""},
		"b01": {"b00", "c01"},
		"b02": {"b01", "a02", "c02", "d01"},
		"b03": {"b02", "d00"},
		"c00": {""},
		"c01": {"c00", "b00"},
		"c02": {"c01", "a02"},
		"c03": {"c02", "a03", "b01", "d02"},
		"d00": {""},
		"d01": {"d00", "b01"},
		"d02": {"d01", "c02"},
		"e00": {"", "d02"},
	}

	if !assert.Equal(t, 5, len(nodes), "node count") {
		return
	}
	if !assert.Equal(t, len(expected), len(named), "event count") {
		return
	}

	checkParents(t, named, expected)
}

func TestDAGtoASCIIschemeRand(t *testing.T) {
	assertar := assert.New(t)

	_, ee := GenEventsByNode(5, 10, 3)
	src := delPeerIndex(ee)

	scheme0, err := DAGtoASCIIscheme(src)
	if !assertar.NoError(err) {
		return
	}

	_, _, names := ASCIIschemeToDAG(scheme0)
	got := delPeerIndex(ee)

	if !assertar.Equal(len(src), len(got), "event count") {
		return
	}

	for _, e0 := range src {
		n := e0.Hash().String()
		e1 := names[n]

		parents0 := edges2text(e0)
		parents1 := edges2text(e1)
		if assertar.EqualValues(parents0, parents1, "at event "+n) {
			continue
		}
		// print info if not EqualValues:
		scheme1, err := DAGtoASCIIscheme(got)
		if !assertar.NoError(err) {
			return
		}
		out := utils.TextColumns(scheme0, scheme1)
		t.Log(out)
		return
	}
}

func TestDAGtoASCIIschemeOptimisation(t *testing.T) {

	t.Run("Simple", func(t *testing.T) {
		testDAGtoASCIIschemeOptimisation(t, `
a00  b00   c00
║    ║    ║║
a01══╣    ║║
║    ║    ║║
╠═══─╫═════c01
║    b01  ╝║
║    ║     ║
a02══╬═════╣
║    ║     ║
║3   ║     ║  // optimise this
║╚═══╬═════c02
║    ║     ║
`, map[string][]string{
			"a00": {""},
			"a01": {"a00", "b00"},
			"a02": {"a01", "b01", "c01"},
			"b00": {""},
			"b01": {"b00", "c00"},
			"c00": {""},
			"c01": {"c00", "a01"},
			"c02": {"c01", "a00", "b01"},
		})
	})

	t.Run("Regression", func(t *testing.T) {
		testDAGtoASCIIschemeOptimisation(t, `
c00    
║       ║      
║       a00    
║       ║       ║      
║       ║       b00    
║       ║       ║      
║       a01═════╣      
║       ║       ║      
c01═════╣       ║      
║║      ║       ║      
║╚═════─╫─══════b01    
║║      ║       ║      
║╚══════a02═════╣      
║      3║       ║ // optimise this
c02════╩╫─══════╣
`, map[string][]string{
			"a00": {""},
			"a01": {"a00", "b00"},
			"a02": {"a01", "b01", "c00"},
			"b00": {""},
			"b01": {"b00", "c00"},
			"c00": {""},
			"c01": {"c00", "a01"},
			"c02": {"c01", "a00", "b01"},
		})
	})

	t.Run("SwapParents", func(t *testing.T) {
		testDAGtoASCIIschemeOptimisation(t, `
a00    
║       ║      
║       b00    
║       ║      
a01═════╣      
║       ║       ║      
║       ║       c00    
║       ║       ║      
║       b01═════╣      
║       ║       ║      
a02═════╬═══════╣      
║║      ║       ║      
║╚═════─╫─══════c01    
║3      ║       ║   // optimise this
║╚══════╬═══════c02
`, map[string][]string{
			"a00": {""},
			"a01": {"a00", "b00"},
			"a02": {"a01", "b01", "c00"},
			"b00": {""},
			"b01": {"b00", "c00"},
			"c00": {""},
			"c01": {"c00", "a01"},
			"c02": {"c01", "a00", "b01"},
		})
	})
}

func testDAGtoASCIIschemeOptimisation(t *testing.T, origScheme string, refs map[string][]string) {
	// step 1: ASCII --> DAG
	_, events, named := ASCIIschemeToDAG(origScheme)
	checkParents(t, named, refs)

	// step 2: DAG --> ASCII
	genScheme, err := DAGtoASCIIscheme(delPeerIndex(events))
	if !assert.NoError(t, err) {
		return
	}

	out := utils.TextColumns(origScheme, genScheme)
	t.Log(out)

	// step 3: ASCII --> DAG (again)
	_, _, named = ASCIIschemeToDAG(genScheme)
	checkParents(t, named, refs)
}

func checkParents(t *testing.T, named map[string]*Event, expected map[string][]string) {
	assertar := assert.New(t)

	for n, e1 := range named {
		parents0 := make(map[string]struct{}, len(expected[n]))
		for _, s := range expected[n] {
			parents0[s] = struct{}{}
		}

		parents1 := make(map[string]struct{}, len(e1.Parents))
		for s := range e1.Parents {
			if s.IsZero() {
				parents1[""] = struct{}{}
			} else {
				parents1[s.String()] = struct{}{}
			}
		}

		if !assertar.Equal(parents0, parents1, "at event "+n) {
			return
		}
	}
}

func edges2text(e *Event) map[string]struct{} {
	res := make(map[string]struct{}, len(e.Parents))
	for p := range e.Parents {
		res[p.String()] = struct{}{}
	}
	return res
}
