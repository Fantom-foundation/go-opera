package fallible

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/kvdb"
	"github.com/Fantom-foundation/go-lachesis/kvdb/memorydb"
)

func TestFallible(t *testing.T) {
	assertar := assert.New(t)

	var (
		key = []byte("test-key")
		val = []byte("test-value")
		db  kvdb.KeyValueStore
		err error
	)

	mem := memorydb.New()
	w := Wrap(mem, nil, nil)
	db = w

	_, err = db.Get(key)
	assertar.NoError(err)

	assertar.Panics(func() {
		db.Put(key, val)
	})

	w.SetWriteCount(1)

	err = db.Put(key, val)
	assertar.NoError(err)

	assertar.Panics(func() {
		err = db.Put(key, val)
	})
}
