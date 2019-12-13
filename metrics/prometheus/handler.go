package prometheus

import (
	"net/http"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var logger = log.New("module", "prometheus")

// ListenTo serves prometheus connections.
func ListenTo(endpoint string, reg metrics.Registry) {
	if reg == nil {
		reg = metrics.DefaultRegistry
	}
	reg.Each(collect)

	go func() {
		logger.Info("metrics server starts", "endpoint", endpoint)
		defer logger.Info("metrics server is stopped")

		http.HandleFunc(
			"/metrics", promhttp.Handler().ServeHTTP)
		err := http.ListenAndServe(endpoint, nil)
		if err != nil {
			logger.Info("metrics server", "err", err)
		}

		// TODO: wait for exit signal?
	}()
}

func collect(name string, metric interface{}) {
	logger.Info("metric to prometheus", "metric", name)

	collector, ok := convertToPrometheusMetric(name, metric)
	if !ok {
		return
	}

	err := prometheus.Register(collector)
	if err != nil {
		switch err.(type) {
		case prometheus.AlreadyRegisteredError:
			return
		default:
			logger.Warn(err.Error())
		}
	}
}
