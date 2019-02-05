package posposet_test

import (
	"math/rand"
	"testing"

	"github.com/Fantom-foundation/go-lachesis/src/posposet"
)

func TestEventHash(t *testing.T) {
	var (
		events = FakeEvents()
		hashes = make([]posposet.EventHash, len(events))
	)

	t.Run("Calculation", func(t *testing.T) {
		for i, e := range events {
			hashes[i] = e.Hash()
		}
	})

	t.Run("Comparison", func(t *testing.T) {
		for i, e := range events {
			hash := e.Hash()
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
