package inter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestASCIIschemeToDAG(t *testing.T) {
	assertar := assert.New(t)

	nodes, _, names := ASCIIschemeToDAG(`
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
╠ ═ b02 ═ ╬   ╣
║   ║     ║  ║║
a03 ╣     ╠ ─ d02
║║  ║     ║  ║║
║║  ║     ║  ║╠ ─ e00
║║  ║     ║   ║   ║
a04 ╫ ─ ─ ╬  ╝║   ║
║║  ║     ║   ║   ║
║╚  ╫╩  ─ c03 ╣   ║
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
		"c00": {""},
		"c01": {"c00", "b00"},
		"c02": {"c01", "a02"},
		"c03": {"c02", "a03", "b01", "d02"},
		"d00": {""},
		"d01": {"d00", "b01"},
		"d02": {"d01", "c02"},
		"e00": {"", "d02"},
	}

	if !assertar.Equal(5, len(nodes), "node count") {
		return
	}
	if !assertar.Equal(len(expected), len(names), "event count") {
		return
	}

	for n, e1 := range names {
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

func TestDAGtoASCIIcheme(t *testing.T) {
	assertar := assert.New(t)

	_, ee := GenEventsByNode(5, 10, 3)
	src := delPeerIndex(ee)

	scheme, err := DAGtoASCIIcheme(src)
	if !assertar.NoError(err) {
		return
	}
	t.Log(scheme)

	_, _, names := ASCIIschemeToDAG(scheme)
	got := delPeerIndex(ee)

	if !assertar.Equal(len(src), len(got), "event count") {
		return
	}

	for _, e0 := range src {
		n := e0.Hash().String()
		e1 := names[n]

		parents0 := make(map[string]struct{}, len(e0.Parents))
		for p := range e0.Parents {
			parents0[p.String()] = struct{}{}
		}

		parents1 := make(map[string]struct{}, len(e1.Parents))
		for p := range e1.Parents {
			parents1[p.String()] = struct{}{}
		}

		if !assertar.EqualValues(parents0, parents1, "at event "+n) {
			return
		}
	}
}
