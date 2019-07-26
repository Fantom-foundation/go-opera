package inter

import (
	"fmt"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter/wire"
)

func TestEventSerialization(t *testing.T) {
	assertar := assert.New(t)

	events := FakeFuzzingEvents()
	for i, e0 := range events {
		dsc := fmt.Sprintf("iter#%d", i)

		w, tt := e0.ToWire()
		w.ExternalTransactions = tt
		buf, err := proto.Marshal(w)
		if !assertar.NoError(err, dsc) {
			break
		}

		w = &wire.Event{}
		err = proto.Unmarshal(buf, w)
		if !assertar.NoError(err, dsc) {
			break
		}
		e1 := WireToEvent(w)

		if !assertar.EqualValues(e0.ExternalTransactions, e1.ExternalTransactions, dsc) ||
			!assertar.EqualValues(e0, e1, dsc) {
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

	t.Run("Hash instead of ExternalTransactions", func(t *testing.T) {
		for _, e := range events {
			h := hash.Of(e.ExternalTransactions.Value...)
			e.ExternalTransactions.Hash = &h
			e.ExternalTransactions.Value = nil
			e.hash = hash.ZeroEvent // drop cache
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
