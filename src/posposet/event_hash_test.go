package posposet_test

import (
	"math/rand"
	"testing"

	"github.com/Fantom-foundation/go-lachesis/src/posposet"
)

func TestEventHash(t *testing.T) {
	creators := []posposet.Address{
		posposet.Address{},
		FakeAddress(),
		FakeAddress(),
		FakeAddress(),
	}
	parents := []posposet.EventHashes{
		nil,
		FakeEventHashes(0),
		FakeEventHashes(1),
		FakeEventHashes(3),
		FakeEventHashes(3),
	}

	var (
		hashes []posposet.EventHash
		events []posposet.Event
	)

	t.Run("Calculation", func(t *testing.T) {
		for c := 0; c < len(creators); c++ {
			for p := 0; p < len(parents); p++ {
				e := posposet.Event{
					Index:   rand.Uint64(),
					Creator: creators[c],
					Parents: parents[p],
				}
				events = append(events, e)
				hashes = append(hashes, e.Hash())
			}
		}
	})

	t.Run("Comparison", func(t *testing.T) {
		for i := 0; i < len(events); i++ {
			hash := events[i].Hash()
			if hash != hashes[i] {
				t.Fatal("Non-deterministic event hash detected")
			}
			for _, other := range hashes[i+1:] {
				if hash == other {
					t.Fatal("Event hash сollision detected")
				}
			}
		}
	})
}

/*
 * Utils:
 */

func FakeEventHash() (h posposet.EventHash) {
	_, err := rand.Read(h[:])
	if err != nil {
		panic(err)
	}
	return
}

func FakeEventHashes(n int) (hh posposet.EventHashes) {
	hh = make(posposet.EventHashes, n)
	for i := 0; i < n; i++ {
		hh[i] = FakeEventHash()
	}
	return
}
