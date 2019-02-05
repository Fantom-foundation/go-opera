package posposet

import (
	"math/rand"
	"testing"
)

func TestAddress(t *testing.T) {
}

/*
 * Utils:
 */

func FakeAddress() (a Address) {
	_, err := rand.Read(a[:])
	if err != nil {
		panic(err)
	}
	return
}
