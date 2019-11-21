package main

import (
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/log"
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

var (
	// TODO: refactor it
	dbDataDirMetric string
	dbSizeMetric    = metrics.NewRegisteredFunctionalGauge("db/size", nil, func() (size int64) {
		if dbDataDirMetric == "" || dbDataDirMetric == "inmemory" {
			return
		}
		err := filepath.Walk(dbDataDirMetric, func(_ string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				size += info.Size()
			}
			return err
		})
		if err != nil {
			log.Error("filepath.Walk", "path", dbDataDirMetric, "err", err)
			return 0
		}
		return
	})
)
