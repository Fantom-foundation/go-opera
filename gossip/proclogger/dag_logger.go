package proclogger

import (
	"time"

	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/logger"
	"github.com/Fantom-foundation/go-opera/utils"
)

func NewLogger() *Logger {
	return &Logger{
		Instance: logger.New(),
	}
}

// EventConnectionStarted starts the event logging
// Not safe for concurrent use
func (l *Logger) EventConnectionStarted(e inter.EventPayloadI, emitted bool) func() {
	l.dagSum.connected++

	start := time.Now()
	l.emitting = emitted
	l.noSummary = true // print summary after the whole event is processed
	l.lastID = e.ID()
	l.lastEventTime = e.CreationTime()

	return func() {
		now := time.Now()
		// logging for the individual item
		msg := "New event"
		logType := l.Log.Debug
		if emitted {
			msg = "New event emitted"
			logType = l.Log.Info
		}
		logType(msg, "id", e.ID(), "parents", len(e.Parents()), "by", e.Creator(),
			"frame", e.Frame(), "txs", e.Txs().Len(),
			"age", utils.PrettyDuration(now.Sub(e.CreationTime().Time())), "t", utils.PrettyDuration(now.Sub(start)))
		// logging for the summary
		l.dagSum.totalProcessing += now.Sub(start)
		l.emitting = false
		l.noSummary = false
		l.summary(now)
	}
}
