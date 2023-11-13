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

		var dbSize int64
		_ = metrics.NewRegisteredFunctionalGauge("db_size", nil, func() int64 {
			return atomic.LoadInt64(&dbSize)
		})

		if !metrics.Enabled {
			return
		}

		if len(datadir) == 0 || datadir == "inmemory" {
			return
		}

		go measureDbDir(datadir, &dbSize)

	})
}

func measureDbDir(datadir string, dbSize *int64) {
	for {
		time.Sleep(10 * time.Second)
		size := sizeOfDir(datadir)
		atomic.StoreInt64(dbSize, size)
	}
}

var (
	symlinksCache     = make(map[string]string, 10e6)
	symlinksThrottler = &throttler{Period: 1000, Timeout: 100 * time.Millisecond}
)

func sizeOfDir(dir string) (size int64) {
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Debug("datadir walk", "path", path, "err", err)
			return filepath.SkipDir
		}

		if info.IsDir() {
			return nil
		}

		dst, cached := symlinksCache[path]
		if !cached {
			symlinksThrottler.Do()
			var err error
			dst, err = filepath.EvalSymlinks(path)
			if err != nil || dst == path {
				dst = ""
			}
			symlinksCache[path] = dst
		}

		if dst != "" {
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
