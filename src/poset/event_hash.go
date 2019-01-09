package poset

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/Fantom-foundation/go-lachesis/src/crypto"
)

type (
	// EventHash is a dedicated type for Event's hash.
	EventHash [sha256.Size]byte

	// EventHashes provides additional methods of EventHash slice.
	EventHashes []EventHash
)

// CalcEventHash returns hash of bytes.
func CalcEventHash(bytes []byte) (hash EventHash) {
	hash.Calc(bytes)
	return
}

// Calc sets value to hash of bytes.
func (hash *EventHash) Calc(bytes []byte) {
	raw := crypto.SHA256(bytes)
	hash.Set(raw)
}

// Set sets value to bytes.
func (hash *EventHash) Set(raw []byte) {
	copy(hash[:], raw)
	for i := len(raw); i < len(hash); i++ {
		hash[i] = 0
	}
}

// Parse sets value to bytes parsed from hex string.
func (hash *EventHash) Parse(raw string) error {
	if raw[0:2] == "0x" {
		raw = raw[2:]
	}
	bytes, err := hex.DecodeString(raw)
	if err != nil {
		return err
	}
	hash.Set(bytes)
	return nil
}

// Equal compares value with bytes.
func (hash *EventHash) Equal(raw []byte) bool {
	if len(hash) != len(raw) {
		return false
	}
	var other EventHash
	other.Set(raw)
	return *hash == other
}

// Bytes returns value as bytes.
func (hash *EventHash) Bytes() []byte {
	var val = *hash
	return val[:]
}

// String returns value as hex string.
func (hash *EventHash) String() string {
	return "0x" + hex.EncodeToString(hash.Bytes())
}

// Zero returns true if zero value.
func (hash *EventHash) Zero() bool {
	for _, b := range hash {
		if b != 0 {
			return false
		}
	}
	return true
}

// Bytes returns values as slice of bytes.
func (hashes EventHashes) Bytes() [][]byte {
	res := make([][]byte, len(hashes))
	for i, hash := range hashes {
		res[i] = hash.Bytes()
	}
	return res
}

// Strings returns values as slice of hex strings.
func (hashes EventHashes) Strings() []string {
	res := make([]string, len(hashes))
	for i, hash := range hashes {
		res[i] = hash.String()
	}
	return res
}

// Contains returns true if there is the hash in values.
func (hashes EventHashes) Contains(hash EventHash) bool {
	for _, h := range hashes {
		if hash == h {
			return true
		}
	}
	return false
}

// Len is a part of sort.Interface implementation.
func (hashes EventHashes) Len() int { return len(hashes) }

// Swap is a part of sort.Interface implementation.
func (hashes EventHashes) Swap(i, j int) { hashes[i], hashes[j] = hashes[j], hashes[i] }

// Less is a part of sort.Interface implementation.
func (hashes EventHashes) Less(i, j int) bool {
	const N = len(EventHash{})
	for n := 0; n < N; n++ {
		if hashes[i][n] != hashes[j][n] {
			return hashes[i][n] < hashes[j][n]
		}
	}
	return false
}
