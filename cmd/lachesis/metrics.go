package main

import (
	"github.com/ethereum/go-ethereum/metrics"
	cli "gopkg.in/urfave/cli.v1"

	"github.com/Fantom-foundation/go-lachesis/metrics/prometheus"
)

var MetricsPrometheusEndpointFlag = cli.StringFlag{
	Name:  "metrics.prometheus.endpoint",
	Usage: "Prometheus API endpoint to report metrics to",
	Value: ":19090",
}

func SetupPrometheus(ctx *cli.Context) {
	if !metrics.Enabled {
		return
	}

	var endpoint = ctx.GlobalString(MetricsPrometheusEndpointFlag.Name)
	prometheus.ListenTo(endpoint, nil)
}
