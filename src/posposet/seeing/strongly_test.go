package seeing

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/ordering"
	"github.com/Fantom-foundation/go-lachesis/src/logger"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/internal"
)

func TestStronglySeen(t *testing.T) {

	t.Run("step 3", func(t *testing.T) {
		testStronglySeen(t, `
a0_1(3)  b0_1     c0_1
║        ║        ║
╠═══════ b1_2     ║
║        ║        ║
║        ╠═══════ c1_3
║        ║        ║
`)
	})

	t.Run("step 4", func(t *testing.T) {
		testStronglySeen(t, `
a0_1(3)  b0_1     c0_1
║        ║        ║
╠═══════ b1_2     ║
║        ║        ║
║        ╠═══════ c1_3
║        ║        ║
║        b2_4 ════╣
║        ║        ║
`)
	})

	t.Run("step 5", func(t *testing.T) {
		testStronglySeen(t, `
a0_1(3)  b0_1(5)  c0_1(5)
║        ║        ║
╠═══════ b1_2(5)  ║
║        ║        ║
║        ╠═══════ c1_3(5)
║        ║        ║
║        b2_4 ════╣
║        ║        ║
a1_5 ════╣        ║
║        ║        ║
`)
	})

}

// testStronglySeen uses event name agreement:
//  "<name>_<level>[(by-level)]",
// where by-level means that event is strongly seen by all event with level >= by-level.
func testStronglySeen(t *testing.T, dag string) {
	logger.SetTestMode(t)
	assertar := assert.New(t)

	peers, _, named := inter.ASCIIschemeToDAG(dag)

	members := make(internal.Members, len(peers))
	for _, peer := range peers {
		members.Add(peer, inter.Stake(1))
	}

	ss := New(members)

	processed := make(map[hash.Event]*inter.Event)
	orderThenProcess := ordering.EventBuffer(ordering.Callback{

		Process: func(e *inter.Event) {
			processed[e.Hash()] = e
			ss.Add(e, 1)
		},

		Drop: func(e *inter.Event, err error) {
			t.Fatal(e, err)
		},

		Exists: func(h hash.Event) *inter.Event {
			return processed[h]
		},
	})

	// push
	for _, e := range named {
		orderThenProcess(e)
	}

	// check
	for dsc, ev := range named {
		_, bylevel := decode(dsc)
		for dsc, by := range named {
			level, _ := decode(dsc)

			who := ss.events[by.Hash()]
			whom := ss.events[ev.Hash()]
			if !assertar.Equal(
				bylevel > 0 && bylevel <= level,
				ss.sufficientCoherence(who, whom),
				fmt.Sprintf("%s strongly sees %s", who.Hash().String(), whom.Hash().String()),
			) {
				return
			}
		}
	}
}

func decode(dsc string) (level, bylevel int64) {
	var err error

	s := strings.Split(dsc, "_")
	s = strings.Split(s[1], "(")

	level, err = strconv.ParseInt(s[0], 10, 32)
	if err != nil {
		panic(err)
	}

	if len(s) < 2 {
		return
	}

	bylevel, err = strconv.ParseInt(strings.TrimRight(s[1], ")"), 10, 32)
	if err != nil {
		panic(err)
	}

	return
}
