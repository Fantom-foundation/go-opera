package vectorindex

import (
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
)

func (vi *Vindex) GetEvent(id hash.Event) *Event {
	event := vi.tempEvents[id]
	if event == nil {
		key := id.Bytes()
		buf, err := vi.eventsDb.Get(key)
		if err != nil {
			vi.Fatal(err)
		}
		if buf == nil {
			return nil
		}

		event = &Event{}
		err = rlp.DecodeBytes(buf, event)
		if err != nil {
			vi.Fatal(err)
		}
	}
	return event
}

func (vi *Vindex) SetEvent(e *Event) {
	key := e.Hash().Bytes()
	buf, err := rlp.EncodeToBytes(e)
	if err != nil {
		vi.Fatal(err)
	}
	err = vi.eventsDb.Put(key, buf)
	if err != nil {
		vi.Fatal(err)
	}
}
