package compactdb

import (
	"io/ioutil"
	"path"
	"testing"

	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/Fantom-foundation/lachesis-base/kvdb/leveldb"
	"github.com/Fantom-foundation/lachesis-base/kvdb/memorydb"
	"github.com/Fantom-foundation/lachesis-base/kvdb/pebble"
	"github.com/stretchr/testify/require"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

func tmpDir(name string) string {
	dir, err := ioutil.TempDir("", name)
	if err != nil {
		panic(err)
	}
	return dir
}

func TestLastKey(t *testing.T) {
	testLastKey(t, memorydb.New())
	dir := tmpDir("test-last-key")
	ldb, err := leveldb.New(path.Join(dir, "ldb"), 16*opt.MiB, 64, nil, nil)
	require.NoError(t, err)
	defer ldb.Close()
	testLastKey(t, ldb)
	pbl, err := pebble.New(path.Join(dir, "pbl"), 16*opt.MiB, 64, nil, nil)
	require.NoError(t, err)
	defer pbl.Close()
	testLastKey(t, pbl)
}

func testLastKey(t *testing.T, db kvdb.Store) {
	require.Nil(t, lastKey(db))

	db.Put([]byte{0}, []byte{0})
	require.Equal(t, []byte{0}, lastKey(db))

	db.Put([]byte{1}, []byte{0})
	require.Equal(t, []byte{1}, lastKey(db))

	db.Put([]byte{2}, []byte{0})
	require.Equal(t, []byte{2}, lastKey(db))

	db.Put([]byte{1, 0}, []byte{0})
	require.Equal(t, []byte{2}, lastKey(db))

	db.Put([]byte{3}, []byte{0})
	require.Equal(t, []byte{3}, lastKey(db))

	db.Put([]byte{3, 0}, []byte{0})
	require.Equal(t, []byte{3, 0}, lastKey(db))

	db.Put([]byte{3, 1}, []byte{0})
	require.Equal(t, []byte{3, 1}, lastKey(db))

	db.Put([]byte{4}, []byte{0})
	require.Equal(t, []byte{4}, lastKey(db))

	db.Put([]byte{4, 0, 0, 0}, []byte{0})
	require.Equal(t, []byte{4, 0, 0, 0}, lastKey(db))

	db.Put([]byte{4, 0, 1, 0}, []byte{0})
	require.Equal(t, []byte{4, 0, 1, 0}, lastKey(db))
}

func TestTrimAfterDiff(t *testing.T) {
	a, b := trimAfterDiff([]byte{}, []byte{}, 1)
	require.Equal(t, []byte{}, a)
	require.Equal(t, []byte{}, b)


	a, b = trimAfterDiff([]byte{1, 2}, []byte{1, 3}, 1)
	require.Equal(t, []byte{1, 2}, a)
	require.Equal(t, []byte{1, 3}, b)

	a, b = trimAfterDiff([]byte{1, 2, 4}, []byte{1, 3, 4}, 1)
	require.Equal(t, []byte{1, 2}, a)
	require.Equal(t, []byte{1, 3}, b)

	a, b = trimAfterDiff([]byte{1, 2, 4, 5}, []byte{1, 3, 4, 6}, 1)
	require.Equal(t, []byte{1, 2}, a)
	require.Equal(t, []byte{1, 3}, b)

	a, b = trimAfterDiff([]byte{1, 2, 4, 5}, []byte{1, 3, 4, 6}, 2)
	require.Equal(t, []byte{1, 2, 4, 5}, a)
	require.Equal(t, []byte{1, 3, 4, 6}, b)

	a, b = trimAfterDiff([]byte{1, 2, 4, 5, 7}, []byte{1, 3, 4, 6}, 2)
	require.Equal(t, []byte{1, 2, 4, 5}, a)
	require.Equal(t, []byte{1, 3, 4, 6}, b)

	a, b = trimAfterDiff([]byte{1, 2, 4, 5, 7}, []byte{1, 3, 4, 6, 7}, 2)
	require.Equal(t, []byte{1, 2, 4, 5}, a)
	require.Equal(t, []byte{1, 3, 4, 6}, b)

	a, b = trimAfterDiff([]byte{1, 2, 4, 5, 7}, []byte{1, 3, 4}, 2)
	require.Equal(t, []byte{1, 2, 4}, a)
	require.Equal(t, []byte{1, 3, 4}, b)
}
