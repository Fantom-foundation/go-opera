package posposet

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/rlp"
)

func TestFrameSerialization(t *testing.T) {
	assert := assert.New(t)
	// fake random data
	nodes, events := GenEventsByNode(4, 10, 3)

	flagTable := FlagTable{}
	cc := eventsByNode{}
	for _, node := range nodes {
		roots := eventsByNode{}
		for _, e := range events[node] {
			roots[e.Creator] = e.Parents
		}
		flagTable[FakeEventHash()] = roots
		if node[0] > 256/2 {
			cc.Add(roots)
		}
	}

	timestamps := timestampsByEvent{
		FakeEventHash(): Timestamp(0),
		FakeEventHash(): Timestamp(rand.Uint64()),
	}

	f0 := &Frame{
		Index:            rand.Uint64(),
		FlagTable:        flagTable,
		ClothoCandidates: cc,
		Atroposes:        timestamps,
		Balances:         common.FakeHash(),
	}
	buf, err := rlp.EncodeToBytes(f0)
	assert.NoError(err)

	f1 := &Frame{}
	err = rlp.DecodeBytes(buf, f1)
	assert.NoError(err)

	assert.EqualValues(f0, f1)
}
