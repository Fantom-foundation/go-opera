package metrics

import (
	"sync/atomic"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/metrics"
	"github.com/minio/minio/pkg/disk"
	cli "gopkg.in/urfave/cli.v1"

	"github.com/Fantom-foundation/go-opera/metrics/prometheus"
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
	prometheus.SetNamespace("opera")
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

// NOTICE THAT walks all files under a directory to calculate size isn't efficient.
func measureDbDir() (size int64) {
	datadir, ok := dbDir.Load().(string)
	if !ok || datadir == "" || datadir == "inmemory" {
		return
	}

	// returned total sizes of the partition which contains datadir.
	// This method is more efficient than files.Walk but may cause inaccuracy.
	di, err := disk.GetInfo(datadir)
	if err != nil {
		log.Error("failed to measure datadir", "path", datadir, "err", err)
		return 0
	}

	return int64(di.Used)
}
