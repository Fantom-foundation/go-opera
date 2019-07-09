package posposet

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
)

// eventsByFrame maps frame num --> roots.
type eventsByFrame map[idx.Frame]EventsByPeer

// Add appends roots of frame.
func (ee eventsByFrame) Add(n idx.Frame, roots EventsByPeer) {
	dest := ee[n]
	if dest == nil {
		dest = EventsByPeer{}
	}
	dest.Add(roots)
	ee[n] = dest
}

// FrameNumsDesc returns sorted frame numbers.
func (ee eventsByFrame) FrameNumsDesc() []idx.Frame {
	var nums []idx.Frame
	for n := range ee {
		nums = append(nums, n)
	}
	sort.Sort(sort.Reverse(frameNums(nums)))
	return nums
}

// String returns human readable string representation.
func (ee eventsByFrame) String() string {
	var ss []string
	for frame, roots := range ee {
		ss = append(ss, fmt.Sprintf("%d: %s", frame, roots.String()))
	}
	return "byFrame{" + strings.Join(ss, ", ") + "}"
}

/*
 * Poset's methods:
 */

// FrameOfEvent returns unfinished frame where event is in.
func (p *Poset) FrameOfEvent(event hash.Event) (frame *Frame, isRoot bool) {
	fnum := p.store.GetEventFrame(event)
	if fnum == nil {
		return
	}

	frame = p.frame(*fnum, false)
	knowns := frame.FlagTable[event]
	for _, events := range knowns {
		if events.Contains(event) {
			isRoot = true
			return
		}
	}

	return
}

func (p *Poset) frameFromStore(n idx.Frame) *Frame {
	if n < p.LastFinishedFrameN() {
		p.Fatalf("too old frame %d is requested", n)
	}
	// return ephemeral
	if n == 0 {
		return &Frame{
			Index:    0,
			Balances: p.Genesis,
		}
	}

	f := p.store.GetFrame(n, p.SuperFrameN)
	if f == nil {
		return p.frameFromStore(n - 1)
	}

	return f
}

// frame finds or creates frame.
func (p *Poset) frame(n idx.Frame, orCreate bool) *Frame {
	if n < p.LastFinishedFrameN() && orCreate {
		p.Fatalf("too old frame %d is requested", n)
	}
	// return ephemeral
	if n == 0 {
		return &Frame{
			Index:    0,
			Balances: p.Genesis,
		}
	}

	// return existing
	f, ok := p.frames[n]
	if !ok {
		if !orCreate {
			return nil
		}
		// create new frame
		f = &Frame{
			Index:     n,
			FlagTable: FlagTable{},
			Balances:  p.frame(n-1, true).Balances,
		}
		p.setFrameSaving(f)
		p.frames[n] = f
		f.save()
	}

	return f
}

// frameNumLast returns last frame number.
func (p *Poset) frameNumLast() idx.Frame {
	var max idx.Frame
	for n := range p.frames {
		if max < n {
			max = n
		}

	}
	return max
}

/*
 * Utils:
 */

type frameNums []idx.Frame

func (p frameNums) Len() int           { return len(p) }
func (p frameNums) Less(i, j int) bool { return p[i] < p[j] }
func (p frameNums) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
