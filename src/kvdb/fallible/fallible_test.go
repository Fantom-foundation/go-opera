package fallible

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/src/kvdb"
	"github.com/Fantom-foundation/go-lachesis/src/kvdb/memorydb"
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
	w := Wrap(mem)
	db = w

	_, err = db.Get(key)
	assertar.NoError(err)

	err = db.Put(key, val)
	assertar.Equal(errWriteLimit, err)

	w.SetWriteCount(1)

	err = db.Put(key, val)
	assertar.NoError(err)

	err = db.Put(key, val)
	assertar.Equal(errWriteLimit, err)
}
