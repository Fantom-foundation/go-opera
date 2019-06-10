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
	totalCalls = metrics.NewRegisteredCounter("total_calls", nil)
)

func main() {
	sig := make(chan os.Signal, 1)
	defer close(sig)

	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	for {
		select {
		case <-sig:
			return
		default:
			log.Printf("total calls = %d", totalCalls.Value())

			currentTime := time.Now().UTC()
			if currentTime.Second() < 45 {
				totalCalls.Inc(2)
				continue
			}

			totalCalls.Dec(1)

			if currentTime.Hour() >= 23 {
				totalCalls.Reset()
			}
		}
	}
}
