package posposet

import (
	"math/rand"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/wire"
)

func TestFrameSerialization(t *testing.T) {
	assertar := assert.New(t)
	// fake random data
	nodes, events := inter.GenEventsByNode(4, 10, 3)

	flagTable := FlagTable{}
	cc := EventsByPeer{}
	for _, node := range nodes {
		roots := EventsByPeer{}
		for _, e := range events[node] {
			roots[e.Creator] = e.Parents
		}
		flagTable[hash.FakeEvent()] = roots
		if node[0] > 256/2 {
			cc.Add(roots)
		}
	}

	timestamps := TimestampsByEvent{
		hash.FakeEvent(): inter.Timestamp(0),
		hash.FakeEvent(): inter.Timestamp(rand.Uint64()),
	}

	f0 := &Frame{
		Index:            rand.Uint64(),
		FlagTable:        flagTable,
		ClothoCandidates: cc,
		Atroposes:        timestamps,
		Balances:         hash.FakeHash(),
	}
	buf, err := proto.Marshal(f0.ToWire())
	assertar.NoError(err)

	w := &wire.Frame{}
	err = proto.Unmarshal(buf, w)
	assertar.NoError(err)

	f1 := WireToFrame(w)

	assertar.EqualValues(f0, f1)
}
