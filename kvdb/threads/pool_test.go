package threads

import (
	"runtime/debug"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	debug.SetMaxThreads(10)
}

func TestThreadPool(t *testing.T) {
	require := require.New(t)

	for name, pool := range map[string]ThreadPool{
		"global": GlobalPool,
		"local":  ThreadPool{},
	} {
		t.Run(name, func(t *testing.T) {
			require.Equal(pool.Cap(), 10)

			got, release := pool.Lock(0)
			require.Equal(got, 0)
			require.Equal(pool.Cap(), 10)
			release(1)
			require.Equal(pool.Cap(), 10)

			got, release = pool.Lock(11)
			require.Equal(got, 10)
			require.Equal(pool.Cap(), 0)
			release(1)
			require.Equal(pool.Cap(), 1)
			release(0)
			require.Equal(pool.Cap(), 10)
		})
	}
}
