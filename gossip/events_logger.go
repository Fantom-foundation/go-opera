package gossip

import (
	"time"

	"github.com/Fantom-foundation/lachesis-base/inter/idx"

	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/logger"
	"github.com/Fantom-foundation/go-opera/utils"
)

type EventsLogger struct {
	connected       idx.Event
	totalProcessing time.Duration
	nextLogging     time.Time

	logger.Instance
}

func NewEventsLogger() *EventsLogger {
	return &EventsLogger{
		Instance: logger.New(),
	}
}

// EventConnectionStarted starts the event logging
// Not safe for concurrent use
func (l *EventsLogger) EventConnectionStarted(e inter.EventPayloadI, emitted bool) func() {
	l.connected++

	start := time.Now()

	return func() {
		now := time.Now()
		// logging for the individual event
		msg := "New event"
		logType := l.Log.Debug
		if emitted {
			msg = "New event emitted"
			logType = l.Log.Info
		}
		logType(msg, "id", e.ID(), "parents", len(e.Parents()), "by", e.Creator(),
			"frame", e.Frame(), "txs", e.Txs().Len(),
			"age", utils.PrettyDuration(now.Sub(e.CreationTime().Time())), "t", utils.PrettyDuration(now.Sub(start)))
		// logging for the events summary
		l.totalProcessing += now.Sub(start)
		if now.After(l.nextLogging) {
			l.Log.Info("New events summary", "new", l.connected, "last_id", e.ID().String(), "age", utils.PrettyDuration(now.Sub(e.CreationTime().Time())), "t", utils.PrettyDuration(l.totalProcessing))
			l.connected = 0
			l.totalProcessing = 0
			l.nextLogging = start.Add(8 * time.Second)
		}
	}
}
