// + build examples

package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Fantom-foundation/go-lachesis/metrics"
)

var (
	totalCount = metrics.RegisterCounter("total_count", nil)
	unixTime   = metrics.RegisterGauge("unix_time", nil)
	dayPart    = metrics.RegisterGaugeFloat64("day_part", nil)
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
			now := time.Now()

			totalCount.Inc(int64(now.Second()))
			if totalCount.Value() > 1000 {
				totalCount.Reset()
			}
			log.Printf("total count = %d", totalCount.Value())

			unixTime.Update(now.Unix())
			log.Printf("unix time = %d", unixTime.Value())

			dayPart.Update(float64(now.Hour()) / 24.0)
			log.Printf("day part = %f", dayPart.Value())
		}
	}
}
