package posnode

import (
	"strconv"
	"strings"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
)

type (
	interval struct {
		from, to idx.Event
	}

	heights map[hash.Peer]interval
)

// String returns human redable representation.
func (hh heights) String() string {
	buf := &strings.Builder{}

	out := func(s string) {
		if _, err := buf.WriteString(s); err != nil {
			panic(err)
		}
	}

	out("[")
	for id, h := range hh {
		out(id.String() + ": ")
		if h.from < 1 {
			out(strconv.FormatUint(uint64(h.to), 10))
		} else {
			out(strconv.FormatUint(uint64(h.from), 10) + "-" + strconv.FormatUint(uint64(h.to), 10))
		}
		out(", ")
	}
	out("]")

	return buf.String()
}

// ToWire only interval.to.
func (hh heights) ToWire() map[string]uint64 {
	w := make(map[string]uint64, len(hh))
	for id, h := range hh {
		w[id.Hex()] = uint64(h.to)
	}

	return w
}

// wireToHeights only interval.to.
func wireToHeights(w map[string]uint64) heights {
	res := make(heights, len(w))
	for hex, h := range w {
		id := hash.HexToPeer(hex)
		res[id] = interval{to: idx.Event(h)}
	}

	return res
}

// Exclude returns heights excluding excepts.
func (hh heights) Exclude(excepts heights) heights {
	diff := make(heights, len(hh))

	for id, h0 := range hh {
		if h1 := excepts[id]; h1.to < h0.to {
			diff[id] = interval{
				from: h1.to,
				to:   h0.to,
			}
		}
	}

	return diff
}

func max(a, b idx.Event) idx.Event {
	if a > b {
		return a
	}
	return b
}
