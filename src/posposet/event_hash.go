package posposet

import (
	"strings"

	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	"github.com/Fantom-foundation/go-lachesis/src/rlp"
)

type (
	// EventHash is a unique identificator of Event.
	// It is a hash of Event.
	EventHash common.Hash

	// EventHashes provides additional methods of EventHash slice.
	EventHashes []EventHash
)

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

// String returns short string representation.
func (hash EventHash) ShortString() string {
	return (common.Hash)(hash).ShortString()
}

// IsZero returns true if hash is empty.
func (hash EventHash) IsZero() bool {
	return hash == EventHash{}
}

// Strings returns values as slice of hex strings.
func (hashes EventHashes) Strings() []string {
	res := make([]string, len(hashes))
	for i, hash := range hashes {
		res[i] = hash.String()
	}
	return res
}

// String returns short string representation.
func (hashes EventHashes) ShortString() string {
	strs := make([]string, len(hashes))
	for i, hash := range hashes {
		strs[i] = hash.ShortString()
	}
	return "[" + strings.Join(strs, ", ") + "]"
}

// Contains returns true if hash is in.
func (hashes EventHashes) Contains(hash EventHash) bool {
	for _, h := range hashes {
		if hash == h {
			return true
		}
	}
	return false
}
