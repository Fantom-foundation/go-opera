package posposet

import (
	"testing"
)

func TestEventHash(t *testing.T) {
	var (
		events = FakeFuzzingEvents()
		hashes = make([]EventHash, len(events))
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
