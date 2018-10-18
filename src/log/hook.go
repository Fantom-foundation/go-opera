// lachesis_log.go
// lachesis hook to logrus
//
package lachesis_log

import (
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type Hook struct {
	mu   sync.RWMutex
	stat [6]int64 // 6 is current value of len(logrus.AllLevels)
	// must be adjusted if changed in future in logrus
	startTime time.Time
}

// NewLocal installs a test hook for a given local logger.
func NewLocal(logger *logrus.Logger, logLevel string) {
	levels := map[string]bool {"debug": true, "error": true, "fatal": true, "panic" : true, "warn": true}
	if _, exist := levels[logLevel]; exist {
		h := new(Hook)
		h.startTime = time.Now()
		logger.Hooks.Add(h)
	}
}

func (t *Hook) Fire(e *logrus.Entry) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if e.Time.Sub(t.startTime).Seconds() > 20 { // 20 seconds in nanoseconds
		// NOTE: we can not use logrus logger here as it seems is not re-entrant
		// and we are inside log entry processing here

		// logrus message format:
		// time="<stamp>" level="<level>" msg="LOGSTAT" debug=0  info=0 warn=5 error=0 fatal=0 panic=0
		fmt.Printf("time=\"" +  e.Time.Format(time.RFC3339) +
			"\" level=debug msg=\"LOGSTAT\" debug=%d info=%d warn=%d error=%d fatal=%d panic=%d\n",
			t.stat[logrus.DebugLevel],
			t.stat[logrus.InfoLevel],
			t.stat[logrus.WarnLevel],
			t.stat[logrus.ErrorLevel],
			t.stat[logrus.FatalLevel],
			t.stat[logrus.PanicLevel])
		t.stat = [6]int64{} // 6 is current value of len(logrus.AllLevels)
		// must be adjusted if changed in future in logrus
		t.startTime = e.Time
	}
	switch e.Level {
	case logrus.PanicLevel:
		fallthrough
	case logrus.FatalLevel:
		fallthrough
	case logrus.ErrorLevel:
		debug.PrintStack()
		fallthrough
	default:
		t.stat[e.Level]++
	}
	return nil
}

func (t *Hook) Levels() []logrus.Level {
	return logrus.AllLevels
}
