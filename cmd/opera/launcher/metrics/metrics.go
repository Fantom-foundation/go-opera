package metrics

import (
	"os"
	"path/filepath"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/metrics"
)

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
