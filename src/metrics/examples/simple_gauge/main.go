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
	unixTime = metrics.NewRegisteredGauge("unix_time", nil)
	dayPart  = metrics.NewRegisteredGaugeFloat64("day_part", nil)
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

			unixTime.Update(currentTime.Unix())
			log.Printf("unix time = %d", unixTime.Value())

			dayPart.Update(float64(currentTime.Hour()) / 24.0)
			log.Printf("day part = %f", dayPart.Value())
		}
	}
}
