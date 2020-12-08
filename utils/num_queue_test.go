package utils

import (
	"math/rand"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNumQueue(t *testing.T) {

	t.Run("Simple", func(t *testing.T) {
		N := uint64(100)
		q := NewNumQueue(0)
		for i := uint64(1); i <= N; i++ {
			var iter sync.WaitGroup
			iter.Add(1)
			go func(i uint64) {
				defer iter.Done()
				q.WaitFor(i)
			}(i)

			q.Done(i)
			iter.Wait()
		}
	})

	t.Run("Random", func(t *testing.T) {
		require := require.New(t)
		N := 100

		q := NewNumQueue(0)
		output := make(chan uint64, 10)
		nums := rand.Perm(N)

		for _, n := range nums {
			go func(n uint64) {
				q.WaitFor(n - 1)
				output <- n
				if n == uint64(N) {
					close(output)
				}
				q.Done(n)

			}(uint64(n + 1))
		}

		var prev uint64
		for got := range output {
			require.Less(prev, got)
			prev = got
		}
	})
}
