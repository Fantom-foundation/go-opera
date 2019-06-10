// +build examples

package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Fantom-foundation/go-lachesis/src/metrics"
)

var (
	totalCount = metrics.NewRegisteredPresetCounter("total_count", nil, 10)
)

func main() {
	sig := make(chan os.Signal, 1)
	defer close(sig)

	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	ticker := time.NewTicker(time.Second)

	for {
		select {
		case <-sig:
			ticker.Stop()
			return
		case <-ticker.C:
			currentTime := time.Now().UTC()

			totalCount.Inc(int64(currentTime.Second()))

			if totalCount.Value() > 50 {
				totalCount.Reset()
			}

			log.Printf("total count = %d", totalCount.Value())
		}
	}
}
