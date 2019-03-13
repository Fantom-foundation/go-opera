package posposet

import (
	"bytes"
	"math/rand"
	"sort"
	"strings"

	"github.com/golang/protobuf/proto"

	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/crypto"
)

type (
	// EventHash is a unique identificator of Event.
	// It is a hash of Event.
	EventHash common.Hash

	// EventHashSlice is a sortable slice of EventHash.
	EventHashSlice []EventHash

	// EventHashes provides additional methods of EventHash index.
	EventHashes map[EventHash]struct{}
)

var (
	ZeroEventHash = EventHash{}
)

/*
 * EventHash methods:
 */

// EventHashOf calcs hash of event.
func EventHashOf(e *Event) EventHash {
	buf, err := proto.Marshal(e.ToWire())
	if err != nil {
		panic(err)
	}
	return EventHash(crypto.Keccak256Hash(buf))
}

// Bytes returns value as byte slice.
func (hash EventHash) Bytes() []byte {
	return (common.Hash)(hash).Bytes()
}

// BytesToEventHash sets b to EventHash.
// If b is larger than len(h), b will be cropped from the left.
func BytesToEventHash(b []byte) EventHash {
	return EventHash(common.BytesToHash(b))
}

// String returns human readable string representation.
func (hash EventHash) String() string {
	if name, ok := EventNameDict[hash]; ok {
		return name
	}
	return (common.Hash)(hash).ShortString()
}

// IsZero returns true if hash is empty.
func (hash *EventHash) IsZero() bool {
	return *hash == EventHash{}
}

/*
 * EventHashes methods:
 */

func newEventHashes(hash ...EventHash) EventHashes {
	hh := EventHashes{}
	hh.Add(hash...)
	return hh
}

// String returns human readable string representation.
func (hh EventHashes) String() string {
	ss := make([]string, 0, len(hh))
	for hash, _ := range hh {
		ss = append(ss, hash.String())
	}
	return "[" + strings.Join(ss, ", ") + "]"
}

// All returns whole index.
func (hh EventHashes) Slice() EventHashSlice {
	arr := make(EventHashSlice, len(hh))
	i := 0
	for h := range hh {
		arr[i] = h
		i++
	}
	return arr
}

// Add appends hash to the index.
func (hh EventHashes) Add(hash ...EventHash) (changed bool) {
	for _, h := range hash {
		if _, ok := hh[h]; !ok {
			hh[h] = struct{}{}
			changed = true
		}
	}
	return
}

// Contains returns true if hash is in.
func (hh EventHashes) Contains(hash EventHash) bool {
	_, ok := hh[hash]
	return ok
}

// ToWire converts to simple slice.
func (hh EventHashes) ToWire() [][]byte {
	var arr EventHashSlice
	for h := range hh {
		arr = append(arr, h)
	}
	sort.Sort(arr)

	return arr.ToWire()
}

// WireToEventHashes converts from simple slice.
func WireToEventHashes(buf [][]byte) EventHashes {
	hh := EventHashes{}
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
func (hh EventHashSlice) ToWire() [][]byte {
	res := make([][]byte, len(hh))
	for i, h := range hh {
		res[i] = h.Bytes()
	}

	return res
}

// WireToEventHashSlice converts from simple slice.
func WireToEventHashSlice(buf [][]byte) EventHashSlice {
	if buf == nil {
		return nil
	}

	hh := make(EventHashSlice, len(buf))
	for i, b := range buf {
		hh[i] = BytesToEventHash(b)
	}

	return hh
}

func (hh EventHashSlice) Len() int      { return len(hh) }
func (hh EventHashSlice) Swap(i, j int) { hh[i], hh[j] = hh[j], hh[i] }
func (hh EventHashSlice) Less(i, j int) bool {
	return bytes.Compare(hh[i].Bytes(), hh[j].Bytes()) < 0
}

/*
 * Utils:
 */

func FakeEventHash() (h EventHash) {
	_, err := rand.Read(h[:])
	if err != nil {
		panic(err)
	}
	return
}

func FakeEventHashes(n int) EventHashes {
	res := EventHashes{}
	for i := 0; i < n; i++ {
		res.Add(FakeEventHash())
	}
	return res
}
