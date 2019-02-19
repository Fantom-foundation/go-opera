package posposet

import (
	"fmt"
	"sort"
	"strings"
)

// TODO: make Roots internal

// Roots maps frame num --> roots.
type Roots map[uint64]Events

// Add appends roots of frame.
func (fs Roots) Add(frameN uint64, roots Events) {
	dest := fs[frameN]
	if dest == nil {
		dest = Events{}
	}
	dest.Add(roots)
	fs[frameN] = dest
}

// FrameNumsDesc returns sorted frame numbers.
func (fs Roots) FrameNumsDesc() []uint64 {
	var nums []uint64
	for n := range fs {
		nums = append(nums, n)
	}
	sort.Sort(sort.Reverse(FrameNumSlice(nums)))
	return nums
}

// String returns human readable string representation.
func (fs Roots) String() string {
	var ss []string
	for frame, roots := range fs {
		ss = append(ss, fmt.Sprintf("%d: %s", frame, roots.String()))
	}
	return "byFrame{" + strings.Join(ss, ", ") + "}"
}

/*
 * Poset's methods:
 */

// FrameOfEvent returns frame number event is in.
func (p *Poset) FrameOfEvent(event EventHash) (frame *Frame, isRoot bool) {
	var nums []uint64
	for n := range p.frames {
		nums = append(nums, n)
	}
	sort.Sort(sort.Reverse(FrameNumSlice(nums))) // TODO: move to Poset index

	for _, n := range nums {
		frame := p.frame(n, false)
		if knownRoots := frame.FlagTable[event]; knownRoots != nil {
			for _, hashes := range knownRoots {
				if hashes.Contains(event) {
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
	if n < p.state.LastFinishedFrameN {
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
			Index:     n,
			FlagTable: FlagTable{},
		}
	}

	f.save = p.saveFuncForFrame(f)
	p.frames[n] = f

	return f
}

/*
 * Utils:
 */

type FrameNumSlice []uint64

func (p FrameNumSlice) Len() int           { return len(p) }
func (p FrameNumSlice) Less(i, j int) bool { return p[i] < p[j] }
func (p FrameNumSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
