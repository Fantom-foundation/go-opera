package poset

import (
	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/common/hexutil"
	"github.com/Fantom-foundation/go-lachesis/src/crypto"
)

type (
	// EventHash is a dedicated type for Event's hash.
	EventHash common.Hash

	// EventHashes provides additional methods of EventHash slice.
	EventHashes []EventHash
)

// CalcEventHash calculates hash of data.
func CalcEventHash(data []byte) EventHash {
	return EventHash(crypto.Keccak256Hash(data))
}

// Set sets value to bytes.
func (hash *EventHash) Set(raw []byte) {
	(*common.Hash)(hash).SetBytes(raw)
}

// Parse sets value to bytes parsed from hex string.
func (hash *EventHash) Parse(raw string) error {
	b, err := hexutil.Decode(raw)
	if err != nil {
		return err
	}
	hash.Set(b)
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
	return (*common.Hash)(hash).Bytes()
}

// String returns value as hex string.
func (hash *EventHash) String() string {
	return (*common.Hash)(hash).String()
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
