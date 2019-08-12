package inter

import (
	"fmt"
	"github.com/ethereum/go-ethereum/rlp"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
)

func TestEventSerialization(t *testing.T) {
	assertar := assert.New(t)

	events := FakeFuzzingEvents()
	for i, e0 := range events {
		dsc := fmt.Sprintf("iter#%d", i)

		e0.ExternalTransactions.Value = [][]byte{{0}, {1}}
		buf, err := rlp.EncodeToBytes(e0)
		if !assertar.NoError(err, dsc) {
			break
		}

		e1 := &Event{}
		err = rlp.DecodeBytes(buf, e1)
		if !assertar.NoError(err, dsc) {
			break
		}

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
