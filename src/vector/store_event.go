package vector

import (
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
)

// GetEvent from DB.
func (vi *Index) GetEvent(id hash.Event) *event {
	key := id.Bytes()
	buf, err := vi.eventsDb.Get(key)
	if err != nil {
		vi.Log.Crit("Failed to get key-value", "err", err)
	}
	if buf == nil {
		return nil
	}

	e := &event{}
	err = rlp.DecodeBytes(buf, e)
	if err != nil {
		vi.Log.Crit("Failed to decode rlp", "err", err)
	}
	return e
}

// SetEvent to DB.
func (vi *Index) SetEvent(e *event) {
	key := e.Hash().Bytes()
	buf, err := rlp.EncodeToBytes(e)
	if err != nil {
		vi.Log.Crit("Failed to encode rlp", "err", err)
	}
	err = vi.eventsDb.Put(key, buf)
	if err != nil {
		vi.Log.Crit("Failed to put key-value", "err", err)
	}
}
