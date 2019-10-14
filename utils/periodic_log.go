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

// Info is timed log.Info
func (l *PeriodicLogger) Info(period time.Duration, msg string, ctx ...interface{}) {
	if time.Since(l.prevLogTime) > period {
		l.Log.Info(msg, ctx...)
		l.prevLogTime = time.Now()
	}
}

// Warn is timed log.Warn
func (l *PeriodicLogger) Warn(period time.Duration, msg string, ctx ...interface{}) {
	if time.Since(l.prevLogTime) > period {
		l.Log.Warn(msg, ctx...)
		l.prevLogTime = time.Now()
	}
}

// Error is timed log.Error
func (l *PeriodicLogger) Error(period time.Duration, msg string, ctx ...interface{}) {
	if time.Since(l.prevLogTime) > period {
		l.Log.Error(msg, ctx...)
		l.prevLogTime = time.Now()
	}
}

// Debug is timed log.Debug
func (l *PeriodicLogger) Debug(period time.Duration, msg string, ctx ...interface{}) {
	if time.Since(l.prevLogTime) > period {
		l.Log.Debug(msg, ctx...)
		l.prevLogTime = time.Now()
	}
}
