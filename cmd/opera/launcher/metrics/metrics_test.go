package metrics

import (
	"testing"
)

func BenchmarkSizeOfDir(b *testing.B) {
	var (
		datadir = "~/.opera"
		size    int64
	)

	symlinksThrottler = new(throttler) // disabled throttling
	size = sizeOfDir(datadir)          // cache warming
	b.ResetTimer()

	for i := 0; i < (b.N * 10); i++ {
		size = sizeOfDir(datadir)
	}
	b.Log(size)
}
