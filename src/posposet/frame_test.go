package posposet

import (
	"math/rand"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/wire"
)

func TestFrameSerialization(t *testing.T) {
	assertar := assert.New(t)
	// fake random data
	nodes, events := inter.GenEventsByNode(4, 10, 3)

	flagTable := FlagTable{}
	for _, node := range nodes {
		roots := EventsByPeer{}
		for _, e := range events[node] {
			roots[e.Creator] = e.Parents
		}
		flagTable[hash.FakeEvent()] = roots
	}

	f0 := &Frame{
		Index:     idx.Frame(rand.Uint64()),
		FlagTable: flagTable,
		Balances:  hash.FakeHash(),
	}
	buf, err := proto.Marshal(f0.ToWire())
	assertar.NoError(err)

	w := &wire.Frame{}
	err = proto.Unmarshal(buf, w)
	assertar.NoError(err)

	f1 := WireToFrame(w)

	assertar.EqualValues(f0, f1)
}
