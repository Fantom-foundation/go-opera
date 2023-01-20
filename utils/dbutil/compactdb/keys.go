package compactdb

import (
	"bytes"
	"math/big"

	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/Fantom-foundation/lachesis-base/kvdb/table"
	"github.com/ethereum/go-ethereum/common"
)

func isEmptyDB(db kvdb.Iteratee) bool {
	it := db.NewIterator(nil, nil)
	defer it.Release()
	return !it.Next()
}

func firstKey(db kvdb.Store) []byte {
	it := db.NewIterator(nil, nil)
	defer it.Release()
	if !it.Next() {
		return nil
	}
	return common.CopyBytes(it.Key())
}

func lastKey(db kvdb.Store) []byte {
	var start []byte
	for {
		for b := 0xff; b >= 0; b-- {
			if !isEmptyDB(table.New(db, append(start, byte(b)))) {
				start = append(start, byte(b))
				break
			}
			if b == 0 {
				return start
			}
		}
	}
}

func keysRange(db kvdb.Store) ([]byte, []byte, *big.Int) {
	first := firstKey(db)
	if first == nil {
		return nil, nil, big.NewInt(0)
	}
	last := lastKey(db)
	if last == nil {
		return nil, nil, big.NewInt(0)
	}
	keySize := len(last)
	if keySize < len(first) {
		keySize = len(first)
	}
	first = common.RightPadBytes(first, keySize)
	last = common.RightPadBytes(last, keySize)
	firstBn := new(big.Int).SetBytes(first)
	lastBn := new(big.Int).SetBytes(last)
	return first, last, new(big.Int).Sub(lastBn, firstBn)
}

func addToKey(prefix *big.Int, diff *big.Int, size int) []byte {
	endBn := new(big.Int).Set(prefix)
	endBn.Add(endBn, diff)
	if len(endBn.Bytes()) > size {
		// overflow
		return bytes.Repeat([]byte{0xff}, size)
	}
	end := endBn.Bytes()
	res := make([]byte, size-len(end), size)
	return append(res, end...)
}

// trimAfterDiff erases all bytes after first *maxDiff* differences between *a* and *b*
func trimAfterDiff(a, b []byte, maxDiff int) ([]byte, []byte) {
	size := len(a)
	if size > len(b) {
		size = len(b)
	}
	for i := 0; i < size; i++ {
		if a[i] != b[i] {
			maxDiff--
			if maxDiff <= 0 {
				size = i + 1
				break
			}
		}
	}
	if len(a) > size {
		a = a[:size]
	}
	if len(b) > size {
		b = b[:size]
	}
	return a, b
}
