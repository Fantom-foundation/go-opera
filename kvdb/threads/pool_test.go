package threads

import (
	"os"
	"runtime/debug"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	debug.SetMaxThreads(10)

	os.Exit(m.Run())
}

func TestThreadPool(t *testing.T) {

	for name, pool := range map[string]ThreadPool{
		"global": GlobalPool,
		"local":  ThreadPool{},
	} {
		t.Run(name, func(t *testing.T) {
			require := require.New(t)

			require.Equal(8, pool.Cap())

			got, release := pool.Lock(0)
			require.Equal(0, got)
			release(1)

			gotA, releaseA := pool.Lock(10)
			require.Equal(8, gotA)
			releaseA(1)

			gotB, releaseB := pool.Lock(10)
			require.Equal(1, gotB)
			releaseB(gotB)

			releaseA(gotA)
			gotB, releaseB = pool.Lock(10)
			require.Equal(8, gotB)

			// don't releaseB(gotB) to check pools isolation
		})
	}
}
