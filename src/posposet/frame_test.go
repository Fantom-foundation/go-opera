package posposet

import (
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
)

func TestFrameSerialization(t *testing.T) {
	assertar := assert.New(t)
	// fake random data
	nodes := inter.GenNodes(4)
	events := inter.GenEventsByNode(nodes, 10, 3, nil, nil)

	roots := EventsByPeer{}
	for _, node := range nodes {
		for _, e := range events[node] {
			roots[e.Creator] = e.Parents.Set()
		}
	}

	f0 := &Frame{
		Index:      idx.Frame(rand.Uint64()),
		Roots:      roots,
		TimeOffset: 3,
		TimeRatio:  1,
	}
	buf, err := rlp.EncodeToBytes(f0)
	assertar.NoError(err)

	f1 := &Frame{}
	err = rlp.DecodeBytes(buf, f1)
	assertar.NoError(err)

	assertar.EqualValues(f0, f1)
}
