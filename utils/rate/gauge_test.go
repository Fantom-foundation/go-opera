package rate

import (
	"sync"
	"testing"
)

func TestGauge_Concurrency(t *testing.T) {
	for try := 0; try < 100; try++ {
		testGaugeConcurrency(t)
	}
}

func testGaugeConcurrency(t *testing.T) {
	g := NewGauge()
	barrier := make(chan struct{})
	wg := sync.WaitGroup{}
	end := int64(32)
	for i := int64(0); i <= end; i++ {
		wg.Add(1)
		turn := i
		go func() {
			defer wg.Done()
			select {
			case <-barrier:
				g.Mark(turn)
				if v := int64(g.rateToGauge(float64(2*end), 1)); v < turn {
					t.Errorf("%d >= %d", v, end)
				}
				return
			}
		}()
	}
	close(barrier)
	wg.Wait()
	if v := int64(g.rateToGauge(float64(2*end), 1)); v != end {
		t.Errorf("%d != %d", v, end)
	}
}
