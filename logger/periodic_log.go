package logger

import (
	"time"
)

// Periodic is the same as logger.Instance, but writes only once in a period
type Periodic struct {
	Instance
	prevLogTime time.Time
}

// Info is timed log.Info
func (l *Periodic) Info(period time.Duration, msg string, ctx ...interface{}) {
	if time.Since(l.prevLogTime) > period {
		l.Log.Info(msg, ctx...)
		l.prevLogTime = time.Now()
	}
}

// Warn is timed log.Warn
func (l *Periodic) Warn(period time.Duration, msg string, ctx ...interface{}) {
	if time.Since(l.prevLogTime) > period {
		l.Log.Warn(msg, ctx...)
		l.prevLogTime = time.Now()
	}
}

// Error is timed log.Error
func (l *Periodic) Error(period time.Duration, msg string, ctx ...interface{}) {
	if time.Since(l.prevLogTime) > period {
		l.Log.Error(msg, ctx...)
		l.prevLogTime = time.Now()
	}
}

// Debug is timed log.Debug
func (l *Periodic) Debug(period time.Duration, msg string, ctx ...interface{}) {
	if time.Since(l.prevLogTime) > period {
		l.Log.Debug(msg, ctx...)
		l.prevLogTime = time.Now()
	}
}

// Trace is timed log.Trace
func (l *Periodic) Trace(period time.Duration, msg string, ctx ...interface{}) {
	if time.Since(l.prevLogTime) > period {
		l.Log.Trace(msg, ctx...)
		l.prevLogTime = time.Now()
	}
}
