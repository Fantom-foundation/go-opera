package prometheus

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/metrics"
)

var logger = log.New("module", "prometheus")

// ListenTo serves prometheus connections.
func ListenTo(endpoint string, reg metrics.Registry) {
	if reg == nil {
		reg = metrics.DefaultRegistry
	}

	go func() {
		log.Info("Starting metrics server", "address", fmt.Sprintf("http://%s/metrics", endpoint))
		defer logger.Info("metrics server is stopped")

		m := http.NewServeMux()
		m.Handle("/metrics", Handler(reg))
		err := http.ListenAndServe(endpoint, m)
		if err != nil {
			logger.Info("metrics server", "err", err)
		}

		// TODO: wait for exit signal?
	}()
}

func Handler(reg metrics.Registry) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Gather and pre-sort the metrics to avoid random listings
		var names []string
		reg.Each(func(name string, i interface{}) {
			names = append(names, name)
		})
		sort.Strings(names)

		// Aggregate all the metris into a Prometheus collector
		c := newCollector()

		for _, name := range names {
			i := reg.Get(name)
			name = prometheusDelims(name)
			switch m := i.(type) {
			case metrics.Counter:
				c.addCounter(name, m.Snapshot())
			case metrics.Gauge:
				c.addGauge(name, m.Snapshot())
			case metrics.GaugeFloat64:
				c.addGaugeFloat64(name, m.Snapshot())
			case metrics.Histogram:
				c.addHistogram(name, m.Snapshot())
			case metrics.Meter:
				c.addMeter(name, m.Snapshot())
			case metrics.Timer:
				c.addTimer(name, m.Snapshot())
			case metrics.ResettingTimer:
				c.addResettingTimer(name, m.Snapshot())
			case metrics.Healthcheck:
				g := metrics.NewFunctionalGauge(func() int64 {
					m.Check()
					if err := m.Error(); err != nil {
						return 0
					}
					return 1
				})
				c.addGauge(name, g)
			default:
				log.Warn("Unknown Prometheus metric type", "type", fmt.Sprintf("%T", i))
			}
		}
		w.Header().Add("Content-Type", "text/plain")
		w.Header().Add("Content-Length", fmt.Sprint(c.buff.Len()))
		w.Write(c.buff.Bytes())
	})
}
