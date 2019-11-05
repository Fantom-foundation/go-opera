package prometheus

import (
	"net/http"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const address = ":19090"

var logger = log.New("module", "prometheus")

func init() {
	if !metrics.Enabled {
		return
	}

	enable()
}

func enable() {
	reg := metrics.DefaultRegistry
	reg.Each(collect)

	go func() {
		logger.Info("metrics server starts", "address", address)
		defer logger.Info("metrics server is stopped")

		http.HandleFunc(
			"/metrics", promhttp.Handler().ServeHTTP)
		http.ListenAndServe(address, nil)

		// TODO: wait for exit signal?
	}()
}

func collect(name string, metric interface{}) {
	collector, ok := convertToPrometheusMetric(name, metric)
	if !ok {
		logger.Debug("metric doesn't support prometheus", "metric", name)
		return
	}

	err := prometheus.Register(collector)
	if err != nil {
		switch err.(type) {
		case prometheus.AlreadyRegisteredError:
			return
		default:
			logger.Error(err.Error())
		}
	}
}
