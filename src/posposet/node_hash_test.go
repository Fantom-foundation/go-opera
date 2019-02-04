package posposet_test

import (
	"math/rand"
	"testing"

	"github.com/Fantom-foundation/go-lachesis/src/posposet"
)

func TestAddress(t *testing.T) {
}

/*
 * Utils:
 */

func FakeAddress() (a posposet.Address) {
	_, err := rand.Read(a[:])
	if err != nil {
		panic(err)
	}
	return
}
