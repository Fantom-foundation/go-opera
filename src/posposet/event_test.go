package posposet_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/src/posposet"
	"github.com/Fantom-foundation/go-lachesis/src/rlp"
)

func TestEventSerialization(t *testing.T) {
	assert := assert.New(t)

	events := FakeEvents()
	for _, e0 := range events {
		buf, err := rlp.EncodeToBytes(e0)
		assert.NoError(err)

		e1 := &posposet.Event{}
		err = rlp.DecodeBytes(buf, e1)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(e0, e1)
	}
}

/*
 * Utils:
 */

func FakeEvents() (res []*posposet.Event) {
	creators := []posposet.Address{
		posposet.Address{},
		FakeAddress(),
		FakeAddress(),
		FakeAddress(),
	}
	parents := []posposet.EventHashes{
		FakeEventHashes(0),
		FakeEventHashes(1),
		FakeEventHashes(8),
	}
	for c := 0; c < len(creators); c++ {
		for p := 0; p < len(parents); p++ {
			e := &posposet.Event{
				Creator: creators[c],
				Parents: parents[p],
			}
			res = append(res, e)
		}
	}
	return
}
