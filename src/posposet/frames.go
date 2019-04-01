package posposet

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
)

// eventsByFrame maps frame num --> roots.
type eventsByFrame map[uint64]EventsByNode

// Add appends roots of frame.
func (ee eventsByFrame) Add(frameN uint64, roots EventsByNode) {
	dest := ee[frameN]
	if dest == nil {
		dest = EventsByNode{}
	}
	dest.Add(roots)
	ee[frameN] = dest
}

// FrameNumsDesc returns sorted frame numbers.
func (ee eventsByFrame) FrameNumsDesc() []uint64 {
	var nums []uint64
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
func (p *Poset) FrameOfEvent(event hash.EventHash) (frame *Frame, isRoot bool) {
	for _, n := range p.frameNumsDesc() {
		frame := p.frame(n, false)
		if knowns := frame.FlagTable[event]; knowns != nil {
			for _, events := range knowns {
				if events.Contains(event) {
					return frame, true
				}
			}
			return frame, false
		}
	}
	return nil, false
}

// frame finds or creates frame.
func (p *Poset) frame(n uint64, orCreate bool) *Frame {
	if n < p.state.LastFinishedFrameN && orCreate {
		panic(fmt.Errorf("Too old frame%d is requested", n))
	}
	// return ephemeral
	if n == 0 {
		return &Frame{
			Index:    0,
			Balances: p.state.Genesis,
		}
	}
	// return existing
	f := p.frames[n]
	if f == nil {
		if !orCreate {
			return nil
		}
		// create new frame
		f = &Frame{
			Index:            n,
			FlagTable:        FlagTable{},
			ClothoCandidates: EventsByNode{},
			Atroposes:        TimestampsByEvent{},
			Balances:         p.frame(n-1, true).Balances,
		}
		p.setFrameSaving(f)
		p.frames[n] = f
		f.save()
	}

	return f
}

// frameNumsAsc returns frame numbers sorted from first to last.
func (p *Poset) frameNumsAsc() []uint64 {
	// TODO: cache sorted
	var nums []uint64
	for n := range p.frames {
		nums = append(nums, n)
	}
	sort.Sort(frameNums(nums))
	return nums
}

// frameNumsDesc returns frame numbers sorted from last to first.
func (p *Poset) frameNumsDesc() []uint64 {
	// TODO: cache sorted
	var nums []uint64
	for n := range p.frames {
		nums = append(nums, n)
	}
	sort.Sort(sort.Reverse(frameNums(nums)))
	return nums
}

// frameNumLast returns last frame number.
func (p *Poset) frameNumLast() uint64 {
	var max uint64
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

type frameNums []uint64

func (p frameNums) Len() int           { return len(p) }
func (p frameNums) Less(i, j int) bool { return p[i] < p[j] }
func (p frameNums) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
