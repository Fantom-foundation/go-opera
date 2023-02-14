package metrics

import (
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/metrics"
)

var (
	dbDir        atomic.Value
	dbSize       atomic.Int64
	dbSizeMetric = metrics.NewRegisteredFunctionalGauge("db_size", nil, func() int64 {
		return dbSize.Load()
	})
)

func SetDataDir(datadir string) {
	was := dbDir.Swap(datadir)
	if was != nil {
		panic("SetDataDir() only once!")
	}
	go measureDbDir()
}

func measureDbDir() {
	for {
		time.Sleep(time.Second)

		datadir, ok := dbDir.Load().(string)
		if !ok || len(datadir) == 0 || datadir == "inmemory" {
			dbSize.Store(0)
		} else {
			size := sizeOfDir(datadir)
			dbSize.Store(size)
		}
	}
}

func sizeOfDir(dir string) (size int64) {
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Debug("datadir walk", "path", path, "err", err)
			return filepath.SkipDir
		}

		if info.IsDir() {
			return nil
		}

		dst, err := filepath.EvalSymlinks(path)
		if err == nil && dst != path {
			size += sizeOfDir(dst)
		} else {
			size += info.Size()
		}

		return nil
	})

	if err != nil {
		log.Debug("datadir walk", "path", dir, "err", err)
	}

	return
}
