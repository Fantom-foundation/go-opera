package metrics

import (
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/metrics"
)

var once sync.Once

func SetDataDir(datadir string) {
	once.Do(func() {
		go measureDbDir("db_size", datadir)
	})
}

func measureDbDir(name, datadir string) {
	var (
		dbSize int64
		gauge  metrics.Gauge
		rescan = (len(datadir) > 0 && datadir != "inmemory")
	)
	for {
		time.Sleep(time.Second)

		if rescan {
			size := sizeOfDir(datadir)
			atomic.StoreInt64(&dbSize, size)
		}

		if gauge == nil {
			gauge = metrics.NewRegisteredFunctionalGauge(name, nil, func() int64 {
				return atomic.LoadInt64(&dbSize)
			})
		}

		if !rescan {
			break
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
