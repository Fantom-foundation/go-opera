package hash

import (
	"bytes"
	"math/big"
	"math/rand"
	"sort"
	"strings"
)

type (
	// Event is a unique identifier of event.
	// It is a hash of Event.
	Event Hash

	// OrderedEvents is a sortable slice of event hash.
	OrderedEvents []Event

	// Events provides additional methods of event hash index.
	Events map[Event]struct{}
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

// String returns human readable string representation.
func (h Event) String() string {
	if name := GetEventName(h); len(name) > 0 {
		return name
	}
	return (Hash)(h).ShortString()
}

// IsZero returns true if hash is empty.
func (h *Event) IsZero() bool {
	return *h == Event{}
}

/*
 * Events methods:
 */

// NewEvents makes event hash index.
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
	for h := range hh {
		ss = append(ss, h.String())
	}
	return "[" + strings.Join(ss, ", ") + "]"
}

// Slice returns whole index as slice.
func (hh Events) Slice() OrderedEvents {
	arr := make(OrderedEvents, len(hh))
	i := 0
	for h := range hh {
		arr[i] = h
		i++
	}
	return arr
}

// Add appends hash to the index.
func (hh Events) Add(hash ...Event) (changed bool) {
	for _, h := range hash {
		if _, ok := hh[h]; !ok {
			hh[h] = struct{}{}
			changed = true
		}
	}
	return
}

// Contains returns true if hash is in.
func (hh Events) Contains(hash Event) bool {
	_, ok := hh[hash]
	return ok
}

// ToWire converts to simple slice.
func (hh Events) ToWire(self Event) [][]byte {
	var (
		head OrderedEvents
		tail OrderedEvents
	)

	for h := range hh {
		if h == self {
			head = append(head, h)
		} else {
			tail = append(tail, h)
		}
	}

	if len(head) != 1 {
		panic("there is no 1 self-parent in Events")
	}

	sort.Sort(tail)

	all := append(head, tail...)
	return all.ToWire()
}

// WireToEventHashes converts from simple slice.
func WireToEventHashes(buf [][]byte) (first Event, all Events) {
	if len(buf) > 0 {
		first = BytesToEvent(buf[0])
	}

	all = Events{}
	for _, b := range buf {
		h := BytesToEvent(b)
		all.Add(h)
	}

	return
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
