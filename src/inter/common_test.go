package inter

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
)

func TestASCIIschemeToDAG(t *testing.T) {
	assertar := assert.New(t)

	nodes, events, names := ASCIIschemeToDAG(`
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

	index := make(map[hash.Event]*Event)
	for _, nodeEvents := range events {
		for _, e := range nodeEvents {
			index[e.Hash()] = e
		}
	}

	for eName, e := range names {
		parents := expected[eName]
		if !assertar.Equal(len(parents), len(e.Parents), "at event "+eName) {
			return
		}
		for _, pName := range parents {
			zeroEvent := hash.ZeroEvent
			if pName != "" {
				zeroEvent = names[pName].Hash()
			}
			if !e.Parents.Contains(zeroEvent) {
				t.Fatalf("%s has no parent %s", eName, pName)
			}
		}
	}
}

// TODO: simplify tests to
// GenEventsByNode() --> ASCIIschemeToDAG() --> DAGtoASCIIcheme() --> compare edges of source and result DAGs.
func TestDAGtoASCIIcheme(t *testing.T) {
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
				d01Hash: {},
			},
			LamportTime: 5,
			hash:        b02Hash,
		},

		&Event{
			Index:   1,
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

	waitResult := `a0                   
║                    
a1                   
║    b0              
a2   ║               
╠----╣               
║    ║    c0         
║    ║    ║          
║    ║    c1         
║    ╠----╣          
║    b1   ║          
║    ╠--╫-╣          
║    ║  ║ c2         
╠----╫--╫-╣          
║    ║  ║ ║   d0     
║    ║  ║ ║   ║      
║    ║  ║ ║   d1     
║    ╠--╫-╫---╣      
║    b2 ║ ║   ║      
║    ╠--╫-╫---╣      
a3   ║  ║ ║   ║      
╠--╫-╝  ║ ║   ║      
a4 ║    ║ ║   ║      
╚--╫----╫-╣   ║      
   ║    ║ ║   d2     
   ║    ║ ╠---╣      
   ║    ║ c3  ║      
   ║    ╚-╣   ║      
   ╚------╣   ║      
          ╚---╣      
              ║   e0 
              ╚---╝  
`

	/* endregion*/

	assert.NotPanics(t, func() {
		resultSchema := DAGtoASCIIcheme(events)
		assert.EqualValues(t, strings.Split(waitResult, "\n"), strings.Split(resultSchema, "\n"))
	})
}

func Test_asciiScheme_AddEvent(t *testing.T) {
	t.Run("nil event", func(t *testing.T) {
		scheme := new(asciiScheme)
		assert.Panics(t, func() {
			scheme.AddEvent("", nil)
		})
	})

	t.Run("event with name", func(t *testing.T) {
		scheme := new(asciiScheme)
		assert.NotPanics(t, func() {
			scheme.AddEvent("test", new(Event))
		})
	})

	t.Run("event without name", func(t *testing.T) {
		scheme := new(asciiScheme)
		assert.NotPanics(t, func() {
			scheme.AddEvent("", new(Event))
		})
	})

	t.Run("scheme has other events", func(t *testing.T) {
		scheme := &asciiScheme{
			graph: [][]string{
				{"first"},
				{"", "second"},
			},
			lengthColumn: 2,
		}

		assert.NotPanics(t, func() {
			event := &Event{
				Index:   1,
				hash:    hash.FakeEvent(),
				Creator: hash.FakePeer(),
			}
			scheme.AddEvent("", event)

			eventName := fmt.Sprintf("%s%d", string(firstNodeName), event.Index-1)
			correctNumberNode := uint64(len(scheme.graph) - 1)

			numberNode, nodeIsFind := scheme.nodes[event.Creator]
			assert.True(t, nodeIsFind)
			assert.Equal(t, correctNumberNode, numberNode)

			nameNode, nameNodeIsFind := scheme.nodesName[event.Creator]
			assert.True(t, nameNodeIsFind)
			assert.Equal(t, firstNodeName, nameNode)

			posEvent, isFindPosEvent := scheme.eventsPosition[event.Hash()]
			assert.True(t, isFindPosEvent)
			assert.EqualValues(t, [2]uint64{2, 2}, posEvent)
			assert.EqualValues(t, []string{"", "", eventName},
				scheme.graph[correctNumberNode])

			assert.Equal(t, uint64(3), scheme.lengthColumn)
			assert.Equal(t, firstNodeName+1, scheme.nextNodeName)
		})
	})
}

func Test_asciiScheme_EventsConnect(t *testing.T) {
	t.Run("parent is zero event", func(t *testing.T) {
		child := hash.FakeEvent()
		parent := hash.ZeroEvent
		aEvent := hash.FakeEvent()
		scheme := asciiScheme{
			graph: [][]string{
				{"a0", ""},
				{"", "b0"},
			},
			eventsPosition: map[hash.Event][2]uint64{
				aEvent: {0, 0},
				child:  {1, 1},
			},
			lengthColumn: 2,
		}

		assert.NotPanics(t, func() {
			scheme.EventsConnect(child, parent)
		})

		assert.Equal(t, uint64(2), scheme.lengthColumn)

		assert.EqualValues(t, []string{"a0", ""}, scheme.graph[0])
		assert.EqualValues(t, []string{"", "b0"}, scheme.graph[1])

		assert.EqualValues(t, [2]uint64{0, 0}, scheme.eventsPosition[aEvent])
		assert.EqualValues(t, [2]uint64{1, 1}, scheme.eventsPosition[child])
		assert.Equal(t, 2, len(scheme.eventsPosition))

	})

	t.Run("nof found event", func(t *testing.T) {
		child := hash.FakeEvent()
		parent := hash.FakeEvent()
		scheme := new(asciiScheme)

		assert.Panics(t, func() {
			scheme.EventsConnect(child, parent)
		})
	})

	t.Run("vertical communication events", func(t *testing.T) {
		child := hash.FakeEvent()
		parent := hash.FakeEvent()
		scheme := &asciiScheme{
			graph: [][]string{
				{"a0", "", "a1"},
			},
			eventsPosition: map[hash.Event][2]uint64{
				child:  {0, 2},
				parent: {0, 0},
			},
			lengthColumn: 3,
		}

		assert.NotPanics(t, func() {
			scheme.EventsConnect(child, parent)
		})

		assert.Equal(t, uint64(3), scheme.lengthColumn)
		assert.EqualValues(t, []string{"a0", "║", "a1"}, scheme.graph[0])
		assert.Equal(t, 1, len(scheme.graph))
	})

	t.Run("incorrectly transferred vertical communication events", func(t *testing.T) {
		child := hash.FakeEvent()
		parent := hash.FakeEvent()
		scheme := &asciiScheme{
			graph: [][]string{
				{"a0", "a1"},
			},
			eventsPosition: map[hash.Event][2]uint64{
				child:  {0, 1},
				parent: {0, 0},
			},
			lengthColumn: 2,
		}

		assert.NotPanics(t, func() {
			scheme.EventsConnect(parent, child)
		})

		assert.Equal(t, uint64(3), scheme.lengthColumn)
		assert.EqualValues(t, []string{"a0", "║", "a1"}, scheme.graph[0])
		assert.Equal(t, 1, len(scheme.graph))
	})

	t.Run("vertical communication events", func(t *testing.T) {
		child := hash.FakeEvent()
		parent := hash.FakeEvent()
		scheme := &asciiScheme{
			graph: make([][]string, 0),
			eventsPosition: map[hash.Event][2]uint64{
				child:  {0, 1},
				parent: {0, 0},
			},
			lengthColumn: 2,
		}

		assert.Panics(t, func() {
			scheme.EventsConnect(child, parent)
		})
	})

	t.Run("vertical communication events with connector on communication", func(t *testing.T) {
		child := hash.FakeEvent()
		parent := hash.FakeEvent()

		connectorOnCommunication := map[string]string{
			"-":  "╫",
			"":   "║",
			"╝":  "╣",
			"╚":  "╠",
			"a1": "a1",
		}

		for connector, waitOnCommunication := range connectorOnCommunication {
			t.Run(fmt.Sprintf("vertical communication events with '%s' on communication", connector),
				func(t *testing.T) {
					scheme := &asciiScheme{
						graph: [][]string{
							{"a0", connector, "a2"},
						},
						eventsPosition: map[hash.Event][2]uint64{
							child:  {0, 2},
							parent: {0, 0},
						},
						lengthColumn: 3,
					}

					assert.NotPanics(t, func() {
						scheme.EventsConnect(child, parent)
					})

					assert.Equal(t, uint64(3), scheme.lengthColumn)
					assert.EqualValues(t, []string{"a0", waitOnCommunication, "a2"}, scheme.graph[0])
					assert.Equal(t, 1, len(scheme.graph))
				})
		}
	})

	t.Run("vertical communication events between other nodes", func(t *testing.T) {
		child := hash.FakeEvent()
		parent := hash.FakeEvent()

		connectorOnCommunication := map[string]struct {
			inputGraph  [][]string
			outputGraph [][]string
		}{
			"on left and right horizontal connectors: type 1": {
				inputGraph: [][]string{
					{"", "-", ""},
					{"a0", "", "a1"},
					{"", "-", ""},
				},
				outputGraph: [][]string{
					{"", "-", ""},
					{"a0", "╫", "a1"},
					{"", "-", ""},
				},
			},
			"on left and right horizontal connectors: type 2": {
				inputGraph: [][]string{
					{"", "╠", ""},
					{"a0", "", "a1"},
					{"", "╣", ""},
				},
				outputGraph: [][]string{
					{"", "╠", ""},
					{"a0", "╫", "a1"},
					{"", "╣", ""},
				},
			},
			"on left and right horizontal connectors: type 3": {
				inputGraph: [][]string{
					{"", "╫", ""},
					{"a0", "", "a1"},
					{"", "╫", ""},
				},
				outputGraph: [][]string{
					{"", "╫", ""},
					{"a0", "╫", "a1"},
					{"", "╫", ""},
				},
			},
			"on left horizontal connector": {
				inputGraph: [][]string{
					{"", "-", ""},
					{"a0", "", "a1"},
					{"", "", ""},
				},
				outputGraph: [][]string{
					{"", "-", ""},
					{"a0", "╣", "a1"},
					{"", "", ""},
				},
			},
			"on right horizontal connector": {
				inputGraph: [][]string{
					{"", "", ""},
					{"a0", "", "a1"},
					{"", "-", ""},
				},
				outputGraph: [][]string{
					{"", "", ""},
					{"a0", "╠", "a1"},
					{"", "-", ""},
				},
			},
		}

		for name, params := range connectorOnCommunication {
			t.Run("vertical communication events between other nodes "+name,
				func(t *testing.T) {
					scheme := &asciiScheme{
						graph: params.inputGraph,
						eventsPosition: map[hash.Event][2]uint64{
							child:  {1, 2},
							parent: {1, 0},
						},
						lengthColumn: 3,
					}

					assert.NotPanics(t, func() {
						scheme.EventsConnect(child, parent)
					})

					assert.EqualValues(t, params.outputGraph, scheme.graph)
				})
		}
	})

	t.Run("horizontal communication events", func(t *testing.T) {
		child := hash.FakeEvent()
		parent := hash.FakeEvent()
		scheme := &asciiScheme{
			graph: [][]string{
				{"a0", "", ""},
				{"", "", "b0"},
			},
			eventsPosition: map[hash.Event][2]uint64{
				child:  {1, 2},
				parent: {0, 0},
			},
			lengthColumn: 3,
		}

		assert.NotPanics(t, func() {
			scheme.EventsConnect(child, parent)
		})

		assert.Equal(t, uint64(4), scheme.lengthColumn)
		assert.EqualValues(t, [][]string{
			{"a0", "║", "║", "╚"},
			{"", "", "", "-"},
			{"", "", "b0", "╝"},
		}, scheme.graph)
	})

	t.Run("horizontal communication events with a round of intermediate event", func(t *testing.T) {
		child := hash.FakeEvent()
		parent := hash.FakeEvent()
		scheme := &asciiScheme{
			graph: [][]string{
				{"a0", "", "a1", "", "", "a2"},
				{"", "", "", "", "", "b0"},
			},
			eventsPosition: map[hash.Event][2]uint64{
				hash.FakeEvent(): {0, 5},
				child:            {1, 5},
				parent:           {0, 2},
			},
			lengthColumn: 6,
		}

		assert.NotPanics(t, func() {
			scheme.EventsConnect(child, parent)
		})

		assert.Equal(t, uint64(7), scheme.lengthColumn)
		assert.EqualValues(t, [][]string{
			{"a0", "", "a1", "", "", "a2", ""},
			{"", "", "", "║", "║", "║", "╚"},
			{"", "", "", "", "", "", "-"},
			{"", "", "", "", "", "b0", "╝"},
		}, scheme.graph)
	})

	t.Run("horizontal communication events with a round of intermediate event and other horizontal connectors", func(t *testing.T) {
		child := hash.FakeEvent()
		parent := hash.FakeEvent()
		scheme := &asciiScheme{
			graph: [][]string{
				{"a0", "║", "a1", "║", "║", "a2", "", ""},
				{"", "", "-", "", "", "", "", ""},
				{"", "", "-", "", "b0", "║", "║", "b1"},
				{"", "c0", "╣", "║", "║", "c1", "", ""},
			},
			eventsPosition: map[hash.Event][2]uint64{
				hash.FakeEvent(): {0, 0},
				parent:           {0, 2},
				hash.FakeEvent(): {0, 5},
				hash.FakeEvent(): {3, 1},
				child:            {3, 5},
				hash.FakeEvent(): {2, 4},
				hash.FakeEvent(): {2, 7},
			},
			lengthColumn: 8,
		}

		assert.NotPanics(t, func() {
			scheme.EventsConnect(child, parent)
		})

		assert.Equal(t, uint64(9), scheme.lengthColumn)
		assert.EqualValues(t, [][]string{
			{"a0", "║", "a1", "║", "║", "a2", "", "", ""},
			{"", "", "", "║", "║", "║", "╚", "", ""},
			{"", "", "-", "", "", "", "-", "", ""},
			{"", "", "-", "", "b0", "║", "╫", "║", "b1"},
			{"", "c0", "╣", "║", "║", "c1", "╝", "", ""},
		}, scheme.graph)
	})
}

func Test_asciiScheme_String(t *testing.T) {
	t.Run("empty scheme", func(t *testing.T) {
		scheme := new(asciiScheme)
		result := scheme.String()
		assert.Equal(t, "", result)
	})

	t.Run("not set events position ascii scheme", func(t *testing.T) {
		scheme := &asciiScheme{
			graph: [][]string{
				{"a0"},
			},
		}
		result := scheme.String()
		assert.Equal(t, "", result)
	})

	t.Run("correct ascii scheme", func(t *testing.T) {
		scheme := &asciiScheme{
			graph: [][]string{
				{"a0", "╠", "╣", "╫", "╝", "╩", "╚"},
			},
			eventsPosition: map[hash.Event][2]uint64{
				hash.FakeEvent(): {0, 0},
			},
			lengthColumn: 7,
		}
		result := scheme.String()
		assert.Equal(t, "a0 \n╠--\n╣  \n╫--\n╝  \n╩--\n╚--\n", result)
	})
}
