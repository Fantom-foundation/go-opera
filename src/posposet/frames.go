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
	sort.Sort(sort.Reverse(orderedFrames(nums)))
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
func (p *Poset) FrameOfEvent(event hash.Event) (frame *Frame) {
	for i := idx.Frame(1); true; i++ {
		frame = p.frame(i, false)
		if frame == nil {
			return
		}
		for e := range frame.Events.Each() {
			if e == event {
				return frame
			}
		}
	}

	return nil
}

func (p *Poset) frameFromStore(n idx.Frame) *Frame {
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
			Index:    n,
			Events:   EventsByPeer{},
			Roots:    EventsByPeer{},
			Balances: p.frame(n-1, true).Balances,
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
 * orderedFrames:
 */

type orderedFrames []idx.Frame

func (ff orderedFrames) Len() int           { return len(ff) }
func (ff orderedFrames) Less(i, j int) bool { return ff[i] < ff[j] }
func (ff orderedFrames) Swap(i, j int)      { ff[i], ff[j] = ff[j], ff[i] }

/*
 * uniqueFrames:
 */

type uniqueFrames map[idx.Frame]struct{}

func (ff *uniqueFrames) Add(n idx.Frame) {
	(*ff)[n] = struct{}{}

}

func (ff uniqueFrames) Asc() orderedFrames {
	res := make(orderedFrames, 0, len(ff))
	for n := range ff {
		res = append(res, n)
	}

	sort.Sort(res)
	return res
}

func (ff uniqueFrames) Desc() orderedFrames {
	res := make(orderedFrames, 0, len(ff))
	for n := range ff {
		res = append(res, n)
	}

	sort.Sort(sort.Reverse(res))
	return res
}
