package metrics

import (
	"os"

	"github.com/fsnotify/fsnotify"
)

// StartFileWatcher starts watching the named file or directory (non-recursively) using metrics.Gauge.
//  - name: metric name
//  - path: path to file of directory
func StartFileWatcher(name, path string) (stop func()) {
	metric := RegisterGauge(name, nil)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		Log.Crit("Failed to make fs watcher", "err", err)
	}

	err = watcher.Add(path)
	if err != nil {
		Log.Crit("Failed to add fs watcher", "err", err)
	}

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
						Log.Crit("Failed to get fs info", "err", err)
					}

					metric.Update(fi.Size())
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				Log.Error("Failed to watch fs", "err", err)
			}
		}
	}()

	stop = func() {
		if err := watcher.Close(); err != nil {
			Log.Error("Failed to close fs watcher", "err", err)
		}
	}

	return
}
