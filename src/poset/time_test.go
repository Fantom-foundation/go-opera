package posposet

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/src/inter"
)

func TestLamportTimeCounter(t *testing.T) {
	assertar := assert.New(t)

	data := map[inter.Timestamp][]inter.Timestamp{
		math.MaxUint64: {},
		3:              {3},
		10:             {9, 10, 10, 11},
		2:              {10, 9, 10, 10, 11, 2, 8, 2, 2, 1, 1},
	}

	for expected, vals := range data {
		counter := timeCounter{}
		for _, t := range vals {
			counter.Add(t)
		}

		actual := counter.MaxMin()
		if !assertar.Equal(expected, actual, "max-min time") {
			break
		}
	}
}
