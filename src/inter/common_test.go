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
	t.Run(`Correct ascii scheme`, func(t *testing.T) {
		/*
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
		*/

		// region setup

		aNodePeer := hash.HexToPeer("a")
		bNodePeer := hash.HexToPeer("b")
		cNodePeer := hash.HexToPeer("c")
		dNodePeer := hash.HexToPeer("d")
		eNodePeer := hash.HexToPeer("e")

		a00Hash := hash.HexToEventHash("a00")
		a01Hash := hash.HexToEventHash("a01")
		a02Hash := hash.HexToEventHash("a02")
		a03Hash := hash.HexToEventHash("a03")
		a04Hash := hash.HexToEventHash("a04")

		b00Hash := hash.HexToEventHash("b00")
		b01Hash := hash.HexToEventHash("b01")
		b02Hash := hash.HexToEventHash("b02")

		c00Hash := hash.HexToEventHash("c00")
		c01Hash := hash.HexToEventHash("c01")
		c02Hash := hash.HexToEventHash("c02")
		c03Hash := hash.HexToEventHash("c03")

		d00Hash := hash.HexToEventHash("d00")
		d01Hash := hash.HexToEventHash("d01")
		d02Hash := hash.HexToEventHash("d02")

		e00Hash := hash.HexToEventHash("e00")

		events := Events{
			&Event{
				Index:   1,
				Creator: aNodePeer,
				Parents: map[hash.Event]struct{}{
					hash.ZeroEvent: {},
				},
				LamportTime: 1,
				hash:        a00Hash,
			},
			&Event{
				Index:   2,
				Creator: aNodePeer,
				Parents: map[hash.Event]struct{}{
					a00Hash: {},
				},
				LamportTime: 2,
				hash:        a01Hash,
			},
			&Event{
				Index:   3,
				Creator: aNodePeer,
				Parents: map[hash.Event]struct{}{
					a01Hash: {},
					b00Hash: {},
				},
				LamportTime: 3,
				hash:        a02Hash,
			},
			&Event{
				Index:   4,
				Creator: aNodePeer,
				Parents: map[hash.Event]struct{}{
					a02Hash: {},
					b02Hash: {},
				},
				LamportTime: 6,
				hash:        a03Hash,
			},
			&Event{
				Index:   5,
				Creator: aNodePeer,
				Parents: map[hash.Event]struct{}{
					a03Hash: {},
					c02Hash: {},
					d01Hash: {},
				},
				LamportTime: 7,
				hash:        a04Hash,
			},

			&Event{
				Index:   1,
				Creator: bNodePeer,
				Parents: map[hash.Event]struct{}{
					hash.ZeroEvent: {},
				},
				LamportTime: 1,
				hash:        b00Hash,
			},
			&Event{
				Index:   2,
				Creator: bNodePeer,
				Parents: map[hash.Event]struct{}{
					b00Hash: {},
					c01Hash: {},
				},
				LamportTime: 3,
				hash:        b01Hash,
			},
			&Event{
				Index:   3,
				Creator: bNodePeer,
				Parents: map[hash.Event]struct{}{
					b01Hash: {},
					a02Hash: {},
					c02Hash: {},
					d01Hash: {},
				},
				LamportTime: 5,
				hash:        b02Hash,
			},

			&Event{
				Index:
				1,
				Creator: cNodePeer,
				Parents: map[hash.Event]struct{}{
					hash.ZeroEvent: {},
				},
				LamportTime: 1,
				hash:        c00Hash,
			},
			&Event{
				Index:   2,
				Creator: cNodePeer,
				Parents: map[hash.Event]struct{}{
					c00Hash: {},
					b00Hash: {},
				},
				LamportTime: 2,
				hash:        c01Hash,
			},
			&Event{
				Index:   3,
				Creator: cNodePeer,
				Parents: map[hash.Event]struct{}{
					c01Hash: {},
					a02Hash: {},
				},
				LamportTime: 4,
				hash:        c02Hash,
			},
			&Event{
				Index:   4,
				Creator: cNodePeer,
				Parents: map[hash.Event]struct{}{
					c02Hash: {},
					d02Hash: {},
					a03Hash: {},
					b01Hash: {},
				},
				LamportTime: 7,
				hash:        c03Hash,
			},

			&Event{
				Index:   1,
				Creator: dNodePeer,
				Parents: map[hash.Event]struct{}{
					hash.ZeroEvent: {},
				},
				LamportTime: 1,
				hash:        d00Hash,
			},
			&Event{
				Index:   2,
				Creator: dNodePeer,
				Parents: map[hash.Event]struct{}{
					d00Hash: {},
					b01Hash: {},
				},
				LamportTime: 4,
				hash:        d01Hash,
			},
			&Event{
				Index:   3,
				Creator: dNodePeer,
				Parents: map[hash.Event]struct{}{
					d01Hash: {},
					c02Hash: {},
				},
				LamportTime: 5,
				hash:        d02Hash,
			},

			&Event{
				Index:   1,
				Creator: eNodePeer,
				Parents: map[hash.Event]struct{}{
					hash.ZeroEvent: {},
					d02Hash:        {},
				},
				LamportTime: 6,
				hash:        e00Hash,
			},
		}

		/* endregion*/

		assert.NotPanics(t, func() {
			resultSchema := CreateSchemaByEvents(events)

			assert.EqualValues(t, resultSchema, "")
		})
	})
}
