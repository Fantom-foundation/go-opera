package node

import (
	"testing"
)

// go test -bench "BenchmarkSmartSelectorNext" -benchmem -run "^$" ./src/node

const (
	fakePeersCount = 50
)

func BenchmarkSmartSelectorNext(b *testing.B) {
	participants1 := fakePeers(fakePeersCount)
	participants2 := clonePeers(participants1)
	participants3 := clonePeers(participants1)

	flagTable1 := fakeFlagTable(participants1)
	flagTable2 := make(map[string]int64, len(flagTable1))
	for k, v := range flagTable1 {
		flagTable2[k] = v
	}

	ss1 := NewSmartPeerSelector(
		participants1,
		fakeAddr(0),
		func() (map[string]int64, error) {
			return flagTable1, nil
		},
	)
	ss2 := NewSmartPeerSelector(
		participants2,
		fakeAddr(0),
		func() (map[string]int64, error) {
			return flagTable2, nil
		},
	)
	rnd := NewRandomPeerSelector(
		participants3,
		fakeAddr(0),
	)

	b.ResetTimer()

	b.Run("legacy Next()", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			p := ss1.Next()
			if p == nil {
				b.Fatal("No next peer")
				break
			}
			ss1.UpdateLast(p.PubKeyHex)
		}
	})

	b.Run("modern Next()", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			p := ss2.Next2()
			if p == nil {
				b.Fatal("No next peer")
				break
			}
			ss2.UpdateLast(p.PubKeyHex)
		}
	})

	b.Run("simple Next()", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			p := rnd.Next()
			if p == nil {
				b.Fatal("No next peer")
				break
			}
			rnd.UpdateLast(p.PubKeyHex)
		}
	})

}
