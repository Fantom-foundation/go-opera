package inter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/utils"
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
		"a00": {},
		"a01": {"a00"},
		"a02": {"a01", "b00"},
		"a03": {"a02", "b02"},
		"a04": {"a03", "c02", "d01"},
		"b00": {},
		"b01": {"b00", "c01"},
		"b02": {"b01", "a02", "c02", "d01"},
		"b03": {"b02", "d00"},
		"c00": {},
		"c01": {"c00", "b00"},
		"c02": {"c01", "a02"},
		"c03": {"c02", "a03", "b01", "d02"},
		"d00": {},
		"d01": {"d00", "b01"},
		"d02": {"d01", "c02"},
		"e00": {"d02"},
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

	nodes := GenNodes(5)
	ee := GenRandEvents(nodes, 10, 3, nil)
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
			"a00": {},
			"a01": {"a00", "b00"},
			"a02": {"a01", "b01", "c01"},
			"b00": {},
			"b01": {"b00", "c00"},
			"c00": {},
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
			"a00": {},
			"a01": {"a00", "b00"},
			"a02": {"a01", "b01", "c00"},
			"b00": {},
			"b01": {"b00", "c00"},
			"c00": {},
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
			"a00": {},
			"a01": {"a00", "b00"},
			"a02": {"a01", "b01", "c00"},
			"b00": {},
			"b01": {"b00", "c00"},
			"c00": {},
			"c01": {"c00", "a01"},
			"c02": {"c01", "a00", "b01"},
		})
	})

	t.Run("LostRefs", func(t *testing.T) {
		testDAGtoASCIIschemeOptimisation(t, `
a000                            
║     ║                         
║     b000                      
║     ║     ║                   
║     ╠════ c000                
║     ║     ║                   
a001══╬═════╣                   
║║    ║     ║     ║             
║╚═══─╫─═══─╫─═══ d000          
║     ║     ║     ║             
║     b001══╬═════╣             
║║    ║║    ║     ║     ║       
║╚═══─╫╩═══─╫─═══─╫─═══ e000    
║     ║     ║     ║     ║       
╠════─╫─═══ c001═─╫─════╣       
║     ║     ║     ║     ║       
╠════─╫─═══─╫─═══ d001══╣       
║     ║     ║     ║     ║       
a002═─╫─════╬═════╣     ║       
║     ║     ║     ║     ║       
║     b002══╬═════╣     ║       
║     ║║    ║     ║     ║       
║     ║╚════╬════─╫─═══ e001    
║     ║     ║     ║     ║       
╠════─╫─═══ c002═─╫─════╣       
║     ║     ║     ║     ║       
a003══╬═════╣     ║     ║       
║     ║     ║     ║     ║       
║     ║     ╠════ d002══╣       
║     ║     ║     ║     ║       
║     b003══╬═════╣     ║       
║     ║     ║     ║     ║       
╠═════╬════ c003  ║     ║       
║     ║     ║     ║     ║       
║     ╠═════╬════ d003  ║       
║     ║     ║     ║     ║       
╠════ b004═─╫─════╣     ║       
║     ║     ║     ║     ║       
╠═════╬════ c004  ║     ║       
║     ║     ║     ║     ║       
a004══╬═════╣     ║     ║     // optimise this
║3    ║     ║     ║3    ║       
║╚═══─╫─═══─╫─═══─╫╩═══ e002    
       

`, map[string][]string{
			"a000": {},
			"a001": {"a000", "b000", "c000"},
			"a002": {"a001", "c001", "d001"},
			"a003": {"a002", "b002", "c002"},
			"a004": {"a003", "b004", "c004"},
			"b000": {},
			"b001": {"b000", "c000", "d000"},
			"b002": {"b001", "c001", "d001"},
			"b003": {"b002", "c002", "d002"},
			"b004": {"a003", "b003", "d003"},
			"c000": {"b000"},
			"c001": {"a001", "c000", "e000"},
			"c002": {"a002", "c001", "e001"},
			"c003": {"a003", "b003", "c002"},
			"c004": {"a003", "b004", "c003"},
			"d000": {"a000"},
			"d001": {"a001", "d000", "e000"},
			"d002": {"c002", "d001", "e001"},
			"e000": {"a000", "b000"},
			"e001": {"b001", "c001", "e000"},
			"e002": {"a002", "d001", "e001"},
			"d003": {"b003", "c003", "d002"},
		})
	})
}

func TestDAGtoASCIIFork(t *testing.T) {
	t.Run("Case: Multiple forks", func(t *testing.T) {
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
        ║╚═════─╫─═════ b01
       ║║       ║       ║
       ╚ c02═══─╫─══════╣ // fork
        ║║      ║       ║
        ║╚═════ a02═════╣
        ║       ║       ║
        c03═════╣       ║
        ║║      ║       ║
        ║╚═════─╫─═════ b02
       ║║       ║       ║
       ╚ c04═══─╫─══════╣ // fork
        ║║      ║       ║
        ║╚═════ a03═════╣
	`, map[string][]string{
			"a00": {},
			"a01": {"a00", "b00"},
			"a02": {"a01", "b01", "c01"},
			"a03": {"c03", "a02", "b02"},
			"b00": {},
			"b01": {"b00", "c00"},
			"b02": {"b01", "c02"},
			"c00": {},
			"c01": {"c00", "a01"},
			"c02": {"c00", "b01"},
			"c03": {"c02", "a02"},
			"c04": {"c02", "b02"},
		})
	})

	t.Run("Case: Mixed events with forks.", func(t *testing.T) {
		testDAGtoASCIIschemeOptimisation(t, `
        b00
        ║       ║
        ║       c00
        ║       ║
        b01═════╣
        ║       ║
        ╠══════ c02
        ║       ║
        b02═════╣
        ║       ║
        ╠══════ c04
        ║       ║       ║
        ║       ║       a00
        ║3      ║       ║
        ║╚═════─╫─═════ a01
        ║      3║       ║
        ║      ╚ c01════╣ // fork
        ║║      ║       ║
        ║╚══════╬══════ a02
        ║      3║       ║
        ║      ╚ c03════╣ // fork
        ║       ║       ║
        ╠═══════╬══════ a03
	`, map[string][]string{
			"a00": {},
			"a01": {"a00", "b00"},
			"a02": {"a01", "b01", "c01"},
			"a03": {"c03", "a02", "b02"},
			"b00": {},
			"b01": {"b00", "c00"},
			"b02": {"b01", "c02"},
			"c00": {},
			"c01": {"c00", "a01"},
			"c02": {"c00", "b01"},
			"c03": {"c02", "a02"},
			"c04": {"c02", "b02"},
		})
	})

	t.Run("Case: Fork the very first event.", func(t *testing.T) {
		// TODO: Currently we have issue with hash collision and this test case will be failed.
		// Remove Skip() after merge additional fix from "try-new-event" branch.
		t.Skip()

		testDAGtoASCIIschemeOptimisation(t, `
        a00     b00
        ║       ║   
       ╚ a10    ╠═════════c00           // fork (a10)
	    ║       ║          ║
        ║╚═════─╫─════════─╫─═══════d00
	   ║║       ║          ║         ║
       ╚ a01════╬══════════╬═════════╣
	    ║       ║          ║         ║
        ║╚═════─╫─════════─╫─═══════d01
        ║       ║          ║         ║
        ║       ║          ║         ║ 
        ║       ║          ║         ║ 
        a02════─╫─════════─╫─════════╣
	`, map[string][]string{
			"a00": {},
			"a10": {},
			"a01": {"a00", "d00"},
			"a02": {"a01", "d01"},
			"b00": {},
			"c00": {"b00"},
			"d00": {"a00"},
			"d01": {"d00", "a10"},
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
		for _, s := range e1.Parents {
			parents1[s.String()] = struct{}{}
		}

		if !assertar.Equal(parents0, parents1, "at event "+n) {
			return
		}
	}
}

func edges2text(e *Event) map[string]struct{} {
	res := make(map[string]struct{}, len(e.Parents))
	for _, p := range e.Parents {
		res[p.String()] = struct{}{}
	}
	return res
}
