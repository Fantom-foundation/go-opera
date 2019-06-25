package metrics

import (
	"os"

	"github.com/fsnotify/fsnotify"

	"github.com/Fantom-foundation/go-lachesis/src/logger"
)

// NewFileWatcher starts watching the named file or directory (non-recursively) using metrics.Gauge.
//  - name: metric name
//  - path: path to file of directory
func NewFileWatcher(name, path string) {
	metric := RegisterGauge(name, nil)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logger.Get().Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					fi, err := os.Stat(path)
					if err != nil {
						logger.Get().Fatal(err)
					}

					metric.Update(fi.Size())
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				logger.Get().Info("error:", err)
			}
		}
	}()

	err = watcher.Add(path)
	if err != nil {
		logger.Get().Fatal(err)
	}
	<-done
}
