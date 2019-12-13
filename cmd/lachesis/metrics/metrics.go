package metrics

import (
	"os"
	"path/filepath"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/metrics"
	cli "gopkg.in/urfave/cli.v1"

	"github.com/Fantom-foundation/go-lachesis/metrics/prometheus"
)

var PrometheusEndpointFlag = cli.StringFlag{
	Name:  "metrics.prometheus.endpoint",
	Usage: "Prometheus API endpoint to report metrics to",
	Value: ":19090",
}

func SetupPrometheus(ctx *cli.Context) {
	if !metrics.Enabled {
		return
	}
	prometheus.SetNamespace("lachesis")
	var endpoint = ctx.GlobalString(PrometheusEndpointFlag.Name)
	prometheus.ListenTo(endpoint, nil)
}

var (
	// TODO: refactor it
	dbDir        atomic.Value
	dbSizeMetric = metrics.NewRegisteredFunctionalGauge("db_size", nil, measureDbDir)
)

func SetDataDir(datadir string) {
	dbDir.Store(datadir)
}

func measureDbDir() (size int64) {
	datadir, ok := dbDir.Load().(string)
	if !ok || datadir == "" || datadir == "inmemory" {
		return
	}

	err := filepath.Walk(datadir, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	if err != nil {
		log.Error("filepath.Walk", "path", datadir, "err", err)
		return 0
	}

	return
}
