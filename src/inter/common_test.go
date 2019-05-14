package inter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
)

func TestParseEvents(t *testing.T) {
	assert := assert.New(t)

	nodes, events, names := ParseEvents(`
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

	if !assert.Equal(5, len(nodes), "node count") {
		return
	}
	if !assert.Equal(len(expected), len(names), "event count") {
		return
	}

	index := make(map[hash.Event]*Event)
	for _, nodeEvents := range events {
		for _, e := range nodeEvents {
			index[e.Hash()] = e
		}
	}

	for eName, e := range names {
		parents := expected[eName]
		if !assert.Equal(len(parents), len(e.Parents), "at event "+eName) {
			return
		}
		for _, pName := range parents {
			hash := hash.ZeroEvent
			if pName != "" {
				hash = names[pName].Hash()
			}
			if !e.Parents.Contains(hash) {
				t.Fatalf("%s has no parent %s", eName, pName)
			}
		}
	}
}

func TestCreateSchemaByEvents(t *testing.T) {

	t.Run(`
ascii scheme:

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
║   ║     ║   ║   ║`, func(t *testing.T) {

		// region setup

		aNodePeer := hash.HexToPeer("a")
		bNodePeer := hash.HexToPeer("b")
		cNodePeer := hash.HexToPeer("c")
		dNodePeer := hash.HexToPeer("d")
		eNodePeer := hash.HexToPeer("e")

		events := make(map[string]*Event)

		events["a00"] = &Event{
			Index:   1,
			Creator: aNodePeer,
			Parents: map[hash.Event]struct{}{
				hash.ZeroEvent: {},
			},
			LamportTime: 1,
			hash:        hash.FakeEvent(),
		}
		events["a01"] = &Event{
			Index:   2,
			Creator: aNodePeer,
			Parents: map[hash.Event]struct{}{
				events["a00"].Hash(): {},
			},
			LamportTime: 2,
			hash:        hash.FakeEvent(),
		}
		events["b00"] = &Event{
			Index:   1,
			Creator: bNodePeer,
			Parents: map[hash.Event]struct{}{
				hash.ZeroEvent: {},
			},
			LamportTime: 1,
			hash:        hash.FakeEvent(),
		}
		events["a02"] = &Event{
			Index:   3,
			Creator: aNodePeer,
			Parents: map[hash.Event]struct{}{
				events["a01"].Hash(): {},
				events["b00"].Hash(): {},
			},
			LamportTime: 3,
			hash:        hash.FakeEvent(),
		}
		events["c00"] = &Event{
			Index:
			1,
			Creator: cNodePeer,
			Parents: map[hash.Event]struct{}{
				hash.ZeroEvent: {},
			},
			LamportTime: 1,
			hash:        hash.FakeEvent(),
		}
		events["c01"] = &Event{
			Index:   2,
			Creator: cNodePeer,
			Parents: map[hash.Event]struct{}{
				events["c00"].Hash(): {},
				events["b00"].Hash(): {},
			},
			LamportTime: 2,
			hash:        hash.FakeEvent(),
		}
		events["b01"] = &Event{
			Index:   2,
			Creator: bNodePeer,
			Parents: map[hash.Event]struct{}{
				events["b00"].Hash(): {},
				events["c01"].Hash(): {},
			},
			LamportTime: 3,
			hash:        hash.FakeEvent(),
		}
		events["c02"] = &Event{
			Index:   3,
			Creator: cNodePeer,
			Parents: map[hash.Event]struct{}{
				events["c01"].Hash(): {},
				events["a02"].Hash(): {},
			},
			LamportTime: 4,
			hash:        hash.FakeEvent(),
		}
		events["d00"] = &Event{
			Index:   1,
			Creator: dNodePeer,
			Parents: map[hash.Event]struct{}{
				hash.ZeroEvent: {},
			},
			LamportTime: 1,
			hash:        hash.FakeEvent(),
		}
		events["d01"] = &Event{
			Index:   2,
			Creator: dNodePeer,
			Parents: map[hash.Event]struct{}{
				events["d00"].Hash(): {},
				events["b01"].Hash(): {},
			},
			LamportTime: 4,
			hash:        hash.FakeEvent(),
		}
		events["b02"] = &Event{
			Index:   3,
			Creator: bNodePeer,
			Parents: map[hash.Event]struct{}{
				events["b01"].Hash(): {},
				events["a02"].Hash(): {},
				events["c02"].Hash(): {},
				events["d01"].Hash(): {},
			},
			LamportTime: 5,
			hash:        hash.FakeEvent(),
		}
		events["a03"] = &Event{
			Index:   4,
			Creator: aNodePeer,
			Parents: map[hash.Event]struct{}{
				events["a02"].Hash(): {},
				events["b02"].Hash(): {},
			},
			LamportTime: 6,
			hash:        hash.FakeEvent(),
		}
		events["a04"] = &Event{
			Index:   5,
			Creator: aNodePeer,
			Parents: map[hash.Event]struct{}{
				events["a03"].Hash(): {},
				events["c02"].Hash(): {},
				events["d01"].Hash(): {},
			},
			LamportTime: 7,
			hash:        hash.FakeEvent(),
		}
		events["d02"] = &Event{
			Index:   3,
			Creator: dNodePeer,
			Parents: map[hash.Event]struct{}{
				events["d01"].Hash(): {},
				events["c02"].Hash(): {},
			},
			LamportTime: 5,
			hash:        hash.FakeEvent(),
		}
		events["c03"] = &Event{
			Index:   4,
			Creator: cNodePeer,
			Parents: map[hash.Event]struct{}{
				events["c02"].Hash(): {},
				events["d02"].Hash(): {},
				events["a03"].Hash(): {},
				events["b01"].Hash(): {},
			},
			LamportTime: 7,
			hash:        hash.FakeEvent(),
		}
		events["e00"] = &Event{
			Index:   1,
			Creator: eNodePeer,
			Parents: map[hash.Event]struct{}{
				hash.ZeroEvent:       {},
				events["d02"].Hash(): {},
			},
			LamportTime: 6,
			hash:        hash.FakeEvent(),
		}

		/* endregion*/

		resultSchema := CreateSchemaByEvents(events)

		assert.EqualValues(t, resultSchema, "")
	})
}
