package posposet

import (
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/src/inter"
)

// NOTE: Atroposes B1 and E1 don't look like Figure22 "An Example of time ordering of event blocks in OPERA chain"
const atroposDAG = `
a01     b01     c01
║       ║       ║
a11 ─ ─ ╬ ─ ─ ─ ╣       d01
║       ║       ║       ║
║       ╠ ─ ─ ─ c11 ─ ─ ╣
║       ║       ║       ║       e01
╠ ─ ─ ─ B12+7 ─ ╣       ║       ║
║       ║       ║       ║       ║
║       ║       ╠ ─ ─ ─ D12 ─ ─ ╣
║       ║       ║       ║       ║
A22 ─ ─ ╫ ─ ─ ─ ╬ ─ ─ ─ ╣       ║
║       ║       ║       ║       ║
╠ ─ ─ ─ ╫ ─ ─ ─ ╫ ─ ─ ─ ╬ ─ ─ ─ E12+6
║       ║       ║       ║       ║
╠ ─ ─ ─ ╫ ─ ─ ─ C22 ─ ─ ╣       ║
║       ║       ║       ║       ║
╠ ─ ─ ─ B23 ─ ─ ╣       ║       ║
║       ║       ║       ║       ║
║       ║       ╠ ─ ─ ─ D23 ─ ─ ╣
║       ║       ║       ║       ║
║       ╠ ─ ─ ─ ╫ ─ ─ ─ ╬ ─ ─ ─ E23
║       ║       ║       ║       ║
A33 ─ ─ ╬ ─ ─ ─ ╣       ║       ║
║       ║       ║       ║       ║
║       ╠ ─ ─ ─ C33     ║       ║
║       ║       ║       ║       ║
╠ ─ ─ ─ b33 ─ ─ ╣       ║       ║
║       ║       ║       ║       ║
a43 ─ ─ ╬ ─ ─ ─ ╣       ║       ║
║║      ║       ║       ║       ║
║║      ╠ ─ ─ ─ C44 ─ ─ ╣       ║
║║      ║       ║       ║       ║
╠╫  ─ ─ B44 ─ ─ ╣       ║       ║
║║      ║       ║       ║       ║
║║      ║       ╠ ─ ─ ─ D34 ─ ─ ╣
║║      ║       ║       ║       ║
A54 ─ ─ ╫ ─ ─ ─ ╬ ─ ─ ─ ╣       ║
║║      ║       ║       ║       ║
║╚  ─ ─ ╫ ─ ─ ─ c54 ─ ─ ╣       ║
║║      ║       ║║      ║       ║
║╚  ─ ─ ╫ ─ ─ ─ c64 ─ ─ ╣       ║
║       ║       ║║      ║       ║
║       ║       ╠ ─ ─ ─ ╬ ─ ─ ─ E34
║       ║       ║║      ║      ║║
╠ ─ ─ ─ ╫ ─ ─ ─ ╬ ─ ─ ─ ╫ ─ ─ ─ E45
║       ║       ║║      ║      ║║
╠ ─ ─ ─ B55 ─ ─ ╣║      ║      ║║
║       ║       ║║      ║      ║║
A65 ─ ─ ╬ ─ ─ ─ ╣║      ║      ║║
║       ║       ║║      ║      ║║
╠ ─ ─ ─ ╫ ─ ─ ─ ╫╩ ─  ─ D45    ║║
║       ║       ║       ║      ║║
║       ╠ ─ ─ ─ C75 ─ ─ ╫ ─ ─  ╝║
║       ║       ║       ║      ║║
╠ ─ ─ ─ b65 ─ ─ ╫ ─ ─ ─ ╫ ─ ─  ╝║
║       ║       ║       ║       ║
║       ║       ╠ ─ ─ ─ ╬ ─ ─ ─ E56
║       ║       ║       ║       ║
`

func TestPosetSimpleAtropos(t *testing.T) {
	t.Run("Original poset", func(t *testing.T) {
		testSpecialNamedAtropos(t, false, atroposDAG)
	})
	t.Run("Restored poset", func(t *testing.T) {
		testSpecialNamedAtropos(t, true, atroposDAG)
	})
}

/*
 * Utils:
 */

// testSpecialNamedAtropos is a general test of Atropos selection.
// Node name means:
// - 1st letter uppercase - node should be root;
// - 2nd number - index by node;
// - 3rd number - frame where node should be in;
// - last "+T" - means Atropos and its consensus time;
func testSpecialNamedAtropos(t *testing.T, tryRestoring bool, asciiScheme string) {
	assert := assert.New(t)
	// init
	nodes, _, names := inter.ParseEvents(asciiScheme)
	p, store, input := FakePoset(nodes)
	// process events
	n := 0
	for _, event := range names {
		input.SetEvent(event)
		p.PushEventSync(event.Hash())
		n++
		if tryRestoring && n == len(names)*2/3 {
			ee := p.incompleteEvents
			p = New(store, input)
			p.Bootstrap()
			for _, e := range ee {
				input.SetEvent(&e.Event)
				p.PushEventSync(e.Hash())
			}
		}
	}
	// check each event
	for name, event := range names {
		// check root
		mustBeRoot := (name == strings.ToUpper(name))
		frame, isRoot := p.FrameOfEvent(event.Hash())
		if !assert.Equal(mustBeRoot, isRoot, name+" is root") {
			t.Log(event.String())
			break
		}
		// check frame
		mustBeFrame, err := strconv.ParseUint(name[2:3], 10, 64)
		if !assert.NoError(err, "name the nodes properly: <UpperCaseForRoot><Index><FrameN>") {
			return
		}
		if !assert.Equal(mustBeFrame, frame.Index, "frame of "+name) {
			break
		}
		// check Atropos time
		mustBeAtropos := len(name) > 4 && name[3:4] == "+"
		consensusTime, isAtropos := frame.Atroposes[event.Hash()]
		if !assert.Equal(mustBeAtropos, isAtropos, "Atropos "+name) {
			break
		}
		if !isAtropos {
			continue
		}
		expectedTime, err := strconv.ParseUint(name[4:5], 10, 64)
		if !assert.NoError(err, "name the Atropos properly: <UpperCaseForRoot><Index><FrameN>+<ConsensusTime>") {
			return
		}
		if !assert.Equal(inter.Timestamp(expectedTime), consensusTime, "Atropos "+name+" consensus time") {
			break
		}
	}
}
