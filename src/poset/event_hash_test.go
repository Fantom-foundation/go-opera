package poset

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEventHash(t *testing.T) {
	assertO := assert.New(t)
	var h EventHash

	arrLongTrim := []byte{1, 2, 3, 4}
	arrLongFull := []byte{1, 2, 3, 4, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0} // len = sha256.Size
	arrShortTrim := []byte{3, 2, 1}
	arrShortFull := []byte{3, 2, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0} // len = sha256.Size

	h.Set(arrLongTrim)
	b1 := h.Bytes()

	h.Set(arrShortTrim)
	b2 := h.Bytes()

	assertO.Equal(arrLongFull, b1)
	assertO.Equal(arrShortFull, b2)
}

func TestEventHashes(t *testing.T) {
	assertO := assert.New(t)

	selfParent := GenRootSelfParent(999)
	otherParent := EventHash{}
	hh := EventHashes{selfParent, otherParent}

	bb := hh.Bytes()
	for i := 0; i < hh.Len(); i++ {
		assertO.Equal(hh[i].Bytes(), bb[i])
	}

	ss := hh.Strings()
	for i := 0; i < hh.Len(); i++ {
		assertO.Equal(hh[i].String(), ss[i])
	}
}
