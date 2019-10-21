package inter

import (
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/hash"
)

func TestEventSerialization(t *testing.T) {
	assertar := assert.New(t)

	events := FakeEventWithOneEpoch()
	for i, e0 := range events {
		dsc := fmt.Sprintf("iter#%d", i)

		buf, err := rlp.EncodeToBytes(e0)
		if !assertar.NoError(err, dsc) {
			break
		}

		assertar.Equal(len(buf), e0.Size())
		assertar.Equal(len(buf), e0.CalcSize())

		e1 := &Event{}
		err = rlp.DecodeBytes(buf, e1)
		if !assertar.NoError(err, dsc) {
			break
		}
		if e1.Sig == nil {
			e1.Sig = []uint8{}
		}

		assertar.Equal(len(buf), e1.CalcSize())
		assertar.Equal(len(buf), e1.Size())

		if !assertar.Equal(e0, e1, dsc) {
			break
		}
	}
}

func TestEventHash(t *testing.T) {
	var (
		events = FakeEventWithOneEpoch()
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
