package posposet

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/src/rlp"
)

func TestEventSerialization(t *testing.T) {
	assert := assert.New(t)

	events := FakeFuzzingEvents()
	for _, e0 := range events {
		buf, err := rlp.EncodeToBytes(e0)
		assert.NoError(err)

		e1 := &Event{}
		err = rlp.DecodeBytes(buf, e1)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(e0, e1)
	}
}
