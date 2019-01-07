package node

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSmartSelectorEmpty(t *testing.T) {
	assert := assert.New(t)

	ss := NewSmartPeerSelector(
		fakePeers(0),
		fakeAddr(0),
		func() (map[string]int64, error) {
			return nil, nil
		},
	)

	assert.Nil(ss.Next())
}

func TestSmartSelectorLocalAddrOnly(t *testing.T) {
	assert := assert.New(t)

	ss := NewSmartPeerSelector(
		fakePeers(1),
		fakeAddr(0),
		func() (map[string]int64, error) {
			return nil, nil
		},
	)

	assert.Nil(ss.Next())
}

func TestSmartSelectorUsed(t *testing.T) {
	assert := assert.New(t)

	ss := NewSmartPeerSelector(
		fakePeers(2),
		fakeAddr(0),
		func() (map[string]int64, error) {
			return nil, nil
		},
	)

	assert.Equal(fakeAddr(1), ss.Next().NetAddr)
	assert.Equal(fakeAddr(1), ss.Next().NetAddr)
	assert.Equal(fakeAddr(1), ss.Next().NetAddr)
}

func TestSmartSelectorFlagged(t *testing.T) {
	assert := assert.New(t)

	fp := fakePeers(3)
	fps := fp.ToPeerSlice()

	ss := NewSmartPeerSelector(
		fp,
		fps[0].NetAddr,
		func() (map[string]int64, error) {
			return map[string]int64{
				fps[2].PubKeyHex: 1,
			}, nil
		},
	)

	assert.Equal(fps[1].NetAddr, ss.Next().NetAddr)
	assert.Equal(fps[1].NetAddr, ss.Next().NetAddr)
	assert.Equal(fps[1].NetAddr, ss.Next().NetAddr)
}

func TestSmartSelectorGeneral(t *testing.T) {
	assert := assert.New(t)

	fp := fakePeers(4)
	fps := fp.ToPeerSlice()

	ss := NewSmartPeerSelector(
		fp,
		fps[3].NetAddr,
		func() (map[string]int64, error) {
			return map[string]int64{
				fps[0].PubKeyHex: 0,
				fps[1].PubKeyHex: 0,
				fps[2].PubKeyHex: 1,
				fps[3].PubKeyHex: 0,
			}, nil
		},
	)

	addresses := []string{fps[0].NetAddr, fps[1].NetAddr}
	assert.Contains(addresses, ss.Next().NetAddr)
	assert.Contains(addresses, ss.Next().NetAddr)
	assert.Contains(addresses, ss.Next().NetAddr)
	assert.Contains(addresses, ss.Next().NetAddr)
}

/*
 * go test -bench "BenchmarkSmartSelectorNext" -benchmem -run "^$" ./src/node
 */

func BenchmarkSmartSelectorNext(b *testing.B) {
	const fakePeersCount = 50

	participants1 := fakePeers(fakePeersCount)
	participants2 := clonePeers(participants1)

	flagTable1 := fakeFlagTable(participants1)

	ss1 := NewSmartPeerSelector(
		participants1,
		fakeAddr(0),
		func() (map[string]int64, error) {
			return flagTable1, nil
		},
	)
	rnd := NewRandomPeerSelector(
		participants2,
		fakeAddr(0),
	)

	b.ResetTimer()

	b.Run("smart Next()", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			p := ss1.Next()
			if p == nil {
				b.Fatal("No next peer")
				break
			}
			ss1.UpdateLast(p.PubKeyHex)
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
