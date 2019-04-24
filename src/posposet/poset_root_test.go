package posposet

import (
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/src/inter"
)

func TestPosetSimpleRoot(t *testing.T) {
	testSpecialNamedRoots(t, `
a01   b01   c01
║     ║     ║
a11 ─ ╬ ─ ─ ╣     d01
║     ║     ║     ║
║     ╠ ─ ─ c11 ─ ╣
║     ║     ║     ║     e01
╠ ─ ─ B12 ─ ╣     ║     ║
║     ║     ║     ║     ║
║     ║     ╠ ─ ─ D12 ─ ╣
║     ║     ║     ║     ║
A22 ─ ╫ ─ ─ ╬ ─ ─ ╣     ║
║     ║     ║     ║     ║
╠ ─ ─ ╫ ─ ─ ╫ ─ ─ ╬ ─ ─ E12
║     ║     ║     ║     ║
╠ ─ ─ ╫ ─ ─ C22 ─ ╣     ║
║     ║     ║     ║     ║
╠ ─ ─ B23 ─ ╣     ║     ║
║     ║     ║     ║     ║
║     ║     ╠ ─ ─ D23 ─ ╣
║     ║     ║     ║     ║
║     ╠ ─ ─ ╫ ─ ─ ╬ ─ ─ E23
║     ║     ║     ║     ║
A33 ─ ╬ ─ ─ ╣     ║     ║
║     ║     ║     ║     ║
║     ╠ ─ ─ C33   ║     ║
║     ║     ║     ║     ║
╠ ─ ─ b33 ─ ╣     ║     ║
║     ║     ║     ║     ║
a43 ─ ╬ ─ ─ ╣     ║     ║
║     ║     ║     ║     ║
║     ╠ ─ ─ C44 ─ ╣     ║
║     ║     ║     ║     ║
╠ ─ ─ B44 ─ ╣     ║     ║
║     ║     ║     ║     ║
║     ║     ╠ ─ ─ D34 ─ ╣
║     ║     ║     ║     ║
A54 ─ ╫ ─ ─ ╬ ─ ─ ╣     ║
║     ║     ║     ║     ║
╠ ─ ─ ╫ ─ ─ c54 ─ ╣     ║
║     ║     ║     ║     ║
║     ║     ╠ ─ ─ ╬ ─ ─ E34
║     ║     ║     ║     ║
`)
}

func TestPosetFarRoot(t *testing.T) {
	testSpecialNamedRoots(t, `
a01   b01   c01
║     ║     ║
a11 ─ ╬ ─ ─ ╣     d01
║     ║     ║     ║
║     ╠ ─ ─ c11 ─ ╣
║     ║     ║     ║
╠ ─ ─ B12 ─ ╣     ║
║     ║     ║     ║
║     ╠ ─ ─ ╬ ─ ─ D12
║     ║     ║     ║
A22 ─ ╫ ─ ─ ╬ ─ ─ ╣
║     ║     ║     ║
╠ ─ ─ ╫ ─ ─ C22 ─ ╣
║     ║     ║     ║
╠ ─ ─ B23 ─ ╣     ║
║     ║     ║     ║
║     ╠ ─ ─ ╬ ─ ─ D23
║     ║     ║     ║
A33 ─ ╬ ─ ─ ╣     ║
║     ║     ║     ║
║     ╠ ─ ─ C33   ║
║     ║     ║     ║
╠ ─ ─ b33 ─ ╣     ║
║     ║     ║     ║
a43 ─ ╬ ─ ─ ╣     ║
║     ║     ║     ║
║     ╠ ─ ─ C44 ─ ╣
║     ║     ║     ║     ║
║     ║     ╠ ─ ─ ╬ ─ ─ E04
║     ║     ║     ║     ║
╠ ─ ─ B44 ─ ╣     ║     ║
║     ║     ║     ║     ║
╠ ─ ─ ╫ ─ ─ ╬ ─ ─ D34   ║
║     ║     ║     ║     ║
A54 ─ ╫ ─ ─ ╬ ─ ─ ╣     ║
║     ║     ║     ║     ║
╠ ─ ─ ╫ ─ ─ c54 ─ ╣     ║
║     ║     ║     ║     ║
║     ╠ ─ ─ ╫ ─ ─ ╬ ─ ─ E15
║     ║     ║     ║     ║
║     ║     ╠ ─ ─ ╬ ─ ─ e25
║     ║     ║     ║     ║
`)
}

/*
 * Utils:
 */

// testSpecialNamedRoots is a general test of root selection.
// Node name means:
// - 1st letter uppercase - node should be root;
// - 2nd number - index by node;
// - 3rd number - frame where node should be in;
func testSpecialNamedRoots(t *testing.T, asciiScheme string) {
	assert := assert.New(t)
	// init
	nodes, _, names := inter.ParseEvents(asciiScheme)
	p, _, input := FakePoset(nodes)
	// process events
	for _, event := range names {
		input.SetEvent(event)
		p.PushEventSync(event.Hash())
	}
	// check each
	for name, event := range names {
		// check root
		mustBeRoot := (name == strings.ToUpper(name))
		frame, isRoot := p.FrameOfEvent(event.Hash())
		if !assert.Equal(mustBeRoot, isRoot, name+" is root") {
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
	}
}
