package posposet

import (
	"bytes"
	"io"
	"sort"
	"strings"

	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	"github.com/Fantom-foundation/go-lachesis/src/rlp"
)

type (
	// EventHash is a unique identificator of Event.
	// It is a hash of Event.
	EventHash common.Hash

	// EventHashes provides additional methods of EventHash index.
	EventHashes struct {
		index map[EventHash]struct{}
	}
)

/*
 * EventHash methods:
 */

func EventHash_Zero() EventHash {
	return EventHash{}
}

// EventHashOf calcs hash of event.
func EventHashOf(e *Event) EventHash {
	buf, err := rlp.EncodeToBytes(e)
	if err != nil {
		panic(err)
	}
	return EventHash(crypto.Keccak256Hash(buf))
}

// Bytes returns value as byte slice.
func (hash EventHash) Bytes() []byte {
	return (common.Hash)(hash).Bytes()
}

// String returns value as hex string.
func (hash EventHash) String() string {
	return (common.Hash)(hash).String()
}

// String returns human readable string representation.
func (hash EventHash) ShortString() string {
	return (common.Hash)(hash).ShortString()
}

// IsZero returns true if hash is empty.
func (hash *EventHash) IsZero() bool {
	return *hash == EventHash{}
}

/*
 * EventHashes methods:
 */

// Strings returns values as slice of hex strings.
func (hh *EventHashes) Strings() []string {
	res := make([]string, 0, len(hh.index))
	if hh.index == nil {
		return res
	}
	for hash, _ := range hh.index {
		res = append(res, hash.String())
	}
	return res
}

// String returns human readable string representation.
func (hh *EventHashes) ShortString() string {
	if hh.index == nil {
		return ""
	}
	strs := make([]string, 0, len(hh.index))
	for hash, _ := range hh.index {
		strs = append(strs, hash.ShortString())
	}
	return "[" + strings.Join(strs, ", ") + "]"
}

// All returns index length.
func (hh *EventHashes) Len() int {
	return len(hh.index)
}

// All returns whole index.
func (hh *EventHashes) All() map[EventHash]struct{} {
	return hh.index
}

// Add appends hash to the index.
func (hh *EventHashes) Add(hash ...EventHash) {
	if hh.index == nil {
		hh.index = make(map[EventHash]struct{})
	}
	for _, h := range hash {
		hh.index[h] = struct{}{}
	}
}

// Contains returns true if hash is in.
func (hh *EventHashes) Contains(hash EventHash) bool {
	if hh.index == nil {
		return false
	}
	_, ok := hh.index[hash]
	return ok
}

// EncodeRLP is a specialized encoder to encode index into array.
func (hh *EventHashes) EncodeRLP(w io.Writer) error {
	var arr []EventHash
	for h := range hh.index {
		arr = append(arr, h)
	}
	sort.Sort(sortableEventHashes(arr))
	return rlp.Encode(w, arr)
}

// DecodeRLP is a specialized decoder to decode index from array.
func (hh *EventHashes) DecodeRLP(s *rlp.Stream) error {
	var arr []EventHash
	err := s.Decode(&arr)
	if err != nil {
		return err
	}
	for _, h := range arr {
		hh.Add(h)
	}
	return nil
}

/*
 * Sorting:
 */

type sortableEventHashes []EventHash

func (hh sortableEventHashes) Len() int      { return len(hh) }
func (hh sortableEventHashes) Swap(i, j int) { hh[i], hh[j] = hh[j], hh[i] }
func (hh sortableEventHashes) Less(i, j int) bool {
	return bytes.Compare(hh[i].Bytes(), hh[j].Bytes()) < 0
}
