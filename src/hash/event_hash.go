package hash

import (
	"bytes"
	"fmt"
	"math/big"
	"math/rand"
	"strings"

	"github.com/Fantom-foundation/go-lachesis/src/common/hexutil"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
)

type (
	// Event is a unique identifier of event.
	// It is a hash of Event.
	Event Hash

	// OrderedEvents is a sortable slice of event hash.
	OrderedEvents []Event

	// OrderedEvents is a slice of event hash.
	Events []Event

	EventsStack []Event

	// EventsSet provides additional methods of event hash index.
	EventsSet map[Event]struct{}
)

var (
	// ZeroEvent is a hash of virtual initial event.
	ZeroEvent = Event{}
)

/*
 * Event methods:
 */

// Bytes returns value as byte slice.
func (h Event) Bytes() []byte {
	return (Hash)(h).Bytes()
}

// Big converts a hash to a big integer.
func (h *Event) Big() *big.Int {
	return (*Hash)(h).Big()
}

// SetBytes converts bytes to event hash.
// If b is larger than len(h), b will be cropped from the left.
func (h *Event) SetBytes(raw []byte) {
	(*Hash)(h).SetBytes(raw)
}

// BytesToEvent converts bytes to event hash.
// If b is larger than len(h), b will be cropped from the left.
func BytesToEvent(b []byte) Event {
	return Event(FromBytes(b))
}

// HexToEventHash sets byte representation of s to hash.
// If b is larger than len(h), b will be cropped from the left.
func HexToEventHash(s string) Event {
	return Event(HexToHash(s))
}

// Hex converts an event hash to a hex string.
func (h Event) Hex() string {
	return Hash(h).Hex()
}

// Lamport returns [4:9] bytes, which store event's Lamport.
func (h Event) Lamport() idx.Lamport {
	return idx.BytesToLamport(h[4:8])
}

// Epoch returns [0:4] bytes, which store event's Epoch.
func (h Event) Epoch() idx.SuperFrame {
	return idx.BytesToEpoch(h[0:4])
}

// String returns human readable string representation.
func (h Event) String() string {
	if name := GetEventName(h); len(name) > 0 {
		return name
	}
	// last bytes, because first are occupied by epoch and lamport
	return fmt.Sprintf("%d:%d:%s...", h.Epoch(), h.Lamport(), hexutil.Encode(h[29:32]))
}

// IsZero returns true if hash is empty.
func (h *Event) IsZero() bool {
	return *h == Event{}
}

/*
 * EventsSet methods:
 */

// NewEventsSet makes event hash index.
func NewEventsSet(h ...Event) EventsSet {
	hh := EventsSet{}
	hh.Add(h...)
	return hh
}

// Copy copies events to a new structure.
func (hh EventsSet) Copy() EventsSet {
	ee := make(EventsSet, len(hh))
	for k, v := range hh {
		ee[k] = v
	}

	return ee
}

// String returns human readable string representation.
func (hh EventsSet) String() string {
	ss := make([]string, 0, len(hh))
	for h := range hh {
		ss = append(ss, h.String())
	}
	return "[" + strings.Join(ss, ", ") + "]"
}

// Slice returns whole index as slice.
func (hh EventsSet) Slice() Events {
	arr := make(Events, len(hh))
	i := 0
	for h := range hh {
		arr[i] = h
		i++
	}
	return arr
}

// Add appends hash to the index.
func (hh EventsSet) Add(hash ...Event) (changed bool) {
	for _, h := range hash {
		if _, ok := hh[h]; !ok {
			hh[h] = struct{}{}
			changed = true
		}
	}
	return
}

// Erase erase hash from the index.
func (hh EventsSet) Erase(hash ...Event) (changed bool) {
	for _, h := range hash {
		if _, ok := hh[h]; ok {
			delete(hh, h)
			changed = true
		}
	}
	return
}

// Contains returns true if hash is in.
func (hh EventsSet) Contains(hash Event) bool {
	_, ok := hh[hash]
	return ok
}

/*
 * Events methods:
 */

// NewEvents makes event hash slice.
func NewEvents(h ...Event) Events {
	hh := Events{}
	hh.Add(h...)
	return hh
}

// Copy copies events to a new structure.
func (hh Events) Copy() Events {
	ee := make(Events, len(hh))
	for k, v := range hh {
		ee[k] = v
	}

	return ee
}

// String returns human readable string representation.
func (hh Events) String() string {
	ss := make([]string, 0, len(hh))
	for _, h := range hh {
		ss = append(ss, h.String())
	}
	return "[" + strings.Join(ss, ", ") + "]"
}

// Set returns whole index as a EventsSet.
func (hh Events) Set() EventsSet {
	set := make(EventsSet, len(hh))
	for _, h := range hh {
		set[h] = struct{}{}
	}
	return set
}

// Add appends hash to the slice.
func (hh *Events) Add(hash ...Event) {
	*hh = append(*hh, hash...)
}

/*
 * EventsStack methods:
 */

func (s *EventsStack) Push(v Event) {
	*s = append(*s, v)
}

func (s *EventsStack) Pop() *Event {
	l := len(*s)
	if l == 0 {
		return nil
	}

	res := &(*s)[l-1]
	*s = (*s)[:l-1]

	return res
}

/*
 * OrderedEvents methods:
 */

func (hh OrderedEvents) String() string {
	buf := &strings.Builder{}

	out := func(s string) {
		if _, err := buf.WriteString(s); err != nil {
			panic(err)
		}
	}

	out("[")
	for _, h := range hh {
		out(h.String() + ", ")
	}
	out("]")

	return buf.String()
}

// ToWire converts to simple slice.
func (hh OrderedEvents) ToWire() [][]byte {
	res := make([][]byte, len(hh))
	for i, h := range hh {
		res[i] = h.Bytes()
	}

	return res
}

// WireToOrderedEvents converts from simple slice.
func WireToOrderedEvents(buf [][]byte) OrderedEvents {
	if buf == nil {
		return nil
	}

	hh := make(OrderedEvents, len(buf))
	for i, b := range buf {
		hh[i] = BytesToEvent(b)
	}

	return hh
}

func (hh OrderedEvents) Len() int      { return len(hh) }
func (hh OrderedEvents) Swap(i, j int) { hh[i], hh[j] = hh[j], hh[i] }
func (hh OrderedEvents) Less(i, j int) bool {
	return bytes.Compare(hh[i].Bytes(), hh[j].Bytes()) < 0
}

/*
 * Utils:
 */

// FakeEvent generates random fake event hash for testing purpose.
func FakeEvent() (h Event) {
	_, err := rand.Read(h[:])
	if err != nil {
		panic(err)
	}
	return
}

// FakeEvents generates random fake event hashes for testing purpose.
func FakeEvents(n int) Events {
	res := Events{}
	for i := 0; i < n; i++ {
		res.Add(FakeEvent())
	}
	return res
}
