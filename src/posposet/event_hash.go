package posposet

import (
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

// String returns value as hex string.
func (hash *EventHash) String() string {
	return (*common.Hash)(hash).String()
}

// Strings returns values as slice of hex strings.
func (hashes EventHashes) Strings() []string {
	res := make([]string, len(hashes))
	for i, hash := range hashes {
		res[i] = hash.String()
	}
	return res
}
