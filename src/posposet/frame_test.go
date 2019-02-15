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

	nodes, events := GenEventsByNode(4, 10, 3)

	flagTable := FlagTable{}
	for _, node := range nodes {
		roots := Events{}
		for _, e := range events[node] {
			roots[e.Creator] = e.Parents
		}
		flagTable[node] = roots
	}

	f0 := &Frame{
		Index:     rand.Uint64(),
		FlagTable: flagTable,
		NonRoots:  Events{},
		Balances:  common.FakeHash(),
	}
	buf, err := rlp.EncodeToBytes(f0)
	assert.NoError(err)

	f1 := &Frame{}
	err = rlp.DecodeBytes(buf, f1)
	assert.NoError(err)

	assert.EqualValues(f0, f1)
}
