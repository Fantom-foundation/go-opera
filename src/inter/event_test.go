package inter

import (
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter/wire"
)

func TestEventSerialization(t *testing.T) {
	assert := assert.New(t)

	events := FakeFuzzingEvents()
	for _, e0 := range events {
		buf, err := proto.Marshal(e0.ToWire())
		assert.NoError(err)

		w := &wire.Event{}
		err = proto.Unmarshal(buf, w)
		if !assert.NoError(err) {
			break
		}
		e1 := WireToEvent(w)

		if !assert.Equal(e0, e1) {
			break
		}
	}
}

func TestEventHash(t *testing.T) {
	var (
		events = FakeFuzzingEvents()
		hashes = make([]hash.Event, len(events))
	)

	t.Run("Calculation", func(t *testing.T) {
		for i, e := range events {
			hashes[i] = e.Hash()
		}
	})

	t.Run("Comparison", func(t *testing.T) {
		for i, e := range events {
			h := e.Hash()
			if h != hashes[i] {
				t.Fatal("Non-deterministic event hash detected")
			}
			for _, other := range hashes[i+1:] {
				if h == other {
					t.Fatal("Event hash collision detected")
				}
			}
		}
	})
}
