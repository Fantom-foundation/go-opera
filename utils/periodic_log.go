package utils

import (
	"github.com/Fantom-foundation/go-lachesis/logger"
	"time"
)

// PeriodicLogger is the same as logger.Instance, but writes only once in a period
type PeriodicLogger struct {
	logger.Instance
	prevLogTime time.Time
}

func (l *PeriodicLogger) Info(period time.Duration, msg string, ctx ...interface{}) {
	if time.Since(l.prevLogTime) > period {
		l.Log.Info(msg, ctx...)
		l.prevLogTime = time.Now()
	}
}

func (l *PeriodicLogger) Warn(period time.Duration, msg string, ctx ...interface{}) {
	if time.Since(l.prevLogTime) > period {
		l.Log.Warn(msg, ctx...)
		l.prevLogTime = time.Now()
	}
}

func (l *PeriodicLogger) Error(period time.Duration, msg string, ctx ...interface{}) {
	if time.Since(l.prevLogTime) > period {
		l.Log.Error(msg, ctx...)
		l.prevLogTime = time.Now()
	}
}
