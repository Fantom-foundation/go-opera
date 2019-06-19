package posposet

import (
	"fmt"
	"sort"
	"strings"

	"github.com/pkg/errors"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
)

// eventsByFrame maps frame num --> roots.
type eventsByFrame map[uint64]EventsByPeer

// Add appends roots of frame.
func (ee eventsByFrame) Add(frameN uint64, roots EventsByPeer) {
	dest := ee[frameN]
	if dest == nil {
		dest = EventsByPeer{}
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

// frames errors
var (
	ErrIncorrectFrameKeyType = errors.New("incorrect type frames key")
)

// frame finds or creates frame.
func (p *Poset) frame(n uint64, orCreate bool) *Frame {
	if n < p.state.LastFinishedFrameN && orCreate {
		p.Fatalf("too old frame %d is requested", n)
	}
	// return ephemeral
	if n == 0 {
		return &Frame{
			Index:    0,
			Balances: p.state.Genesis,
		}
	}

	// return existing
	f, ok := p.frames.Load(n)
	if !ok {
		if !orCreate {
			return nil
		}
		// create new frame
		newFrame := &Frame{
			Index:            n,
			FlagTable:        FlagTable{},
			ClothoCandidates: EventsByPeer{},
			Atroposes:        TimestampsByEvent{},
			Balances:         p.frame(n-1, true).Balances,
		}
		p.setFrameSaving(newFrame)
		p.frames.Store(n, newFrame)
		newFrame.save()
		return newFrame
	}

	return f.(*Frame)
}

// frameNumLast returns last frame number.
func (p *Poset) frameNumLast() uint64 {
	var max uint64
	p.frames.Range(func(key, value interface{}) bool {
		n, ok := key.(uint64)
		if !ok {
			p.Fatal(ErrIncorrectFrameKeyType)
		}

		if max < n {
			max = n
		}

		return true
	})

	return max
}

func (p *Poset) mustFrameLoad(key uint64) *Frame {
	f, ok := p.frames.Load(key)
	if !ok {
		p.Fatal(errors.Errorf("frame[%d] doesn't exist", key))
	}

	frame, ok := f.(*Frame)
	if !ok {
		p.Fatal(errors.New("incorrect type frame"))
	}

	return frame
}

/*
 * Utils:
 */

type frameNums []uint64

func (p frameNums) Len() int           { return len(p) }
func (p frameNums) Less(i, j int) bool { return p[i] < p[j] }
func (p frameNums) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
