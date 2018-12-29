package poset

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/Fantom-foundation/go-lachesis/src/crypto"
)

type (
	EventHash   [sha256.Size]byte
	EventHashes []EventHash
)

func CalcEventHash(bytes []byte) (hash EventHash) {
	hash.Calc(bytes)
	return
}

func (hash *EventHash) Calc(bytes []byte) {
	raw := crypto.SHA256(bytes)
	hash.Set(raw)
}

func (hash *EventHash) Set(raw []byte) {
	copy(hash[:], raw)
	for i := len(raw); i < len(hash); i++ {
		hash[i] = 0
	}
}

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

func (hash *EventHash) Equal(raw []byte) bool {
	if len(hash) != len(raw) {
		return false
	}
	var other EventHash
	other.Set(raw)
	return *hash == other
}

func (hash *EventHash) Bytes() []byte {
	var val EventHash = *hash
	return val[:]
}

func (hash *EventHash) String() string {
	return "0x" + hex.EncodeToString(hash.Bytes())
}

func (hash *EventHash) Zero() bool {
	for _, b := range hash {
		if b != 0 {
			return false
		}
	}
	return true
}

func (hashes EventHashes) Bytes() [][]byte {
	res := make([][]byte, len(hashes))
	for i, hash := range hashes {
		res[i] = hash.Bytes()
	}
	return res
}

func (hashes EventHashes) Strings() []string {
	res := make([]string, len(hashes))
	for i, hash := range hashes {
		res[i] = hash.String()
	}
	return res
}

func (hashes EventHashes) Contains(hash EventHash) bool {
	for _, h := range hashes {
		if hash == h {
			return true
		}
	}
	return false
}

func (hashes EventHashes) Len() int      { return len(hashes) }
func (hashes EventHashes) Swap(i, j int) { hashes[i], hashes[j] = hashes[j], hashes[i] }
func (hashes EventHashes) Less(i, j int) bool {
	const N = len(EventHash{})
	for n := 0; n < N; n++ {
		if hashes[i][n] != hashes[j][n] {
			return hashes[i][n] < hashes[j][n]
		}
	}
	return false
}
