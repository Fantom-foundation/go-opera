package hash

import (
	"bytes"
	"math/rand"
	"sort"
	"strings"
)

type (
	// Event is a unique identificator of event.
	// It is a hash of Event.
	Event Hash

	// EventsSlice is a sortable slice of event hash.
	EventsSlice []Event

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

// SetBytes converts bytes to event hash.
// If b is larger than len(h), b will be cropped from the left.
func (h *Event) SetBytes(raw []byte) {
	(*Hash)(h).SetBytes(raw)
}

// BytesToEventHash converts bytes to event hash.
// If b is larger than len(h), b will be cropped from the left.
func BytesToEventHash(b []byte) Event {
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
	if name, ok := EventNameDict[h]; ok {
		return name
	}
	return (Hash)(h).ShortString()
}

// IsZero returns true if hash is empty.
func (h *Event) IsZero() bool {
	return *h == Event{}
}

/*
 * EventHashes methods:
 */

// New Events makes event hash index.
func NewEvents(h ...Event) Events {
	hh := Events{}
	hh.Add(h...)
	return hh
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
func (hh Events) Slice() EventsSlice {
	arr := make(EventsSlice, len(hh))
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
func (hh Events) ToWire() [][]byte {
	var arr EventsSlice
	for h := range hh {
		arr = append(arr, h)
	}
	sort.Sort(arr)

	return arr.ToWire()
}

// WireToEventHashes converts from simple slice.
func WireToEventHashes(buf [][]byte) Events {
	hh := Events{}
	for _, b := range buf {
		h := BytesToEventHash(b)
		hh.Add(h)
	}

	return hh
}

/*
 * EventHashSlice's methods:
 */

// ToWire converts to simple slice.
func (hh EventsSlice) ToWire() [][]byte {
	res := make([][]byte, len(hh))
	for i, h := range hh {
		res[i] = h.Bytes()
	}

	return res
}

// WireToEventHashSlice converts from simple slice.
func WireToEventHashSlice(buf [][]byte) EventsSlice {
	if buf == nil {
		return nil
	}

	hh := make(EventsSlice, len(buf))
	for i, b := range buf {
		hh[i] = BytesToEventHash(b)
	}

	return hh
}

func (hh EventsSlice) Len() int      { return len(hh) }
func (hh EventsSlice) Swap(i, j int) { hh[i], hh[j] = hh[j], hh[i] }
func (hh EventsSlice) Less(i, j int) bool {
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
