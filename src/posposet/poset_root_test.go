package posposet

import (
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	//"github.com/Fantom-foundation/go-lachesis/src/logger"
)

func TestPosetSimpleRoot(t *testing.T) {
	testSpecialNamedRoots(t, `
A101  B101  C101  D101  // 1
║     ║     ║     ║
║     ╠─────╫──── d102
║     ║     ║     ║
║     b102 ─╫─────╣
║     ║     ║     ║
║     ╠─────╫──── d103
a102 ─╣     ║     ║
║     ║     ║     ║
║     b103 ─╣     ║
║     ║     ║     ║
║     ╠─────╫──── d104
║     ║     ║     ║
║     ╠──── c102  ║
║     ║     ║     ║
║     b104 ─╫─────╣
║     ║     ║     ║     // 2
╠─────╫─────╫──── D205
║     ║     ║     ║
A203 ─╫─────╫─────╣
║     ║     ║     ║
a204 ─╫─────╣     ║
║     ║     ║     ║
║     B205 ─╫─────╣
║     ║     ║     ║
║     ╠─────╫──── d206
a205 ─╣     ║     ║
║     ║     ║     ║
╠─────╫──── C203  ║
║     ║     ║     ║
╠─────╫─────╫──── d207
║     ║     ║     ║
╠──── b206  ║     ║
║     ║     ║     ║     // 3
║     B307 ─╫─────╣
║     ║     ║     ║
A306 ─╣     ║     ║
║     ╠─────╫──── D308
║     ║     ║     ║
║     ║     ╠──── d309
╠──── b308  ║     ║
║     ║     ║     ║
╠──── b309  ║     ║
║     ║     C304 ─╣
a307 ─╣     ║     ║
║     ║     ║     ║
║     b310 ─╫─────╣
║     ║     ║     ║
a308 ─╣     ║     ║
║     ╠─────╫──── d310
║     ║     ║     ║
╠──── b311  ║     ║     // 4
║     ║     ╠──── D411
║     ║     ║     ║
║     B412 ─╫─────╣
║     ║     ║     ║
`)
}

/*
 * Utils:
 */

// testSpecialNamedRoots is a general test of root selection.
// Event name means:
// - 1st letter uppercase - event should be root;
// - 2nd number - frame where event should be in;
// - other numbers - index by node (optional);
func testSpecialNamedRoots(t *testing.T, asciiScheme string) {
	//logger.SetTestMode(t)
	assertar := assert.New(t)
	// init
	nodes, _, names := inter.ASCIIschemeToDAG(asciiScheme)
	p, _, input := FakePoset(nodes)
	// process events
	for _, event := range names {
		input.SetEvent(event)
		p.PushEventSync(event.Hash())
	}
	// check each
	for name, event := range names {
		mustBeFrame, mustBeRoot := parseEvent(name)
		// check root
		frame := p.FrameOfEvent(event.Hash())
		isRoot := frame.Roots.Contains(event.Creator, event.Hash())
		if !assertar.Equal(mustBeRoot, isRoot, name+" is root") {
			break
		}
		// check frame
		if !assertar.Equal(idx.Frame(mustBeFrame), frame.Index, "frame of "+name) {
			break
		}
	}
}

func parseEvent(name string) (frameN idx.Frame, isRoot bool) {
	n, err := strconv.ParseUint(name[1:2], 10, 64)
	if err != nil {
		panic(err.Error() + ". Name event " + name + " properly: <UpperCaseForRoot><FrameN><Index>")
	}
	frameN = idx.Frame(n)

	isRoot = name == strings.ToUpper(name)
	return
}
