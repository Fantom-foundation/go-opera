package logger

import (
	"github.com/ethereum/go-ethereum/log"
	"github.com/evalphobia/logrus_sentry"
)

// init with defaults.
func init() {
	log.Root().SetHandler(
		log.CallerStackHandler("%v", log.StdoutHandler))
}

// SetDSN appends sentry hook to log root handler.
func SetDSN(value string) {
	// If DSN is empty, we don't create new hook.
	// Otherwise we'll the same error message for each new log.
	if value == "" {
		log.Warn("Sentry client DSN is empty")
		return
	}

	// TODO: find or make sentry log.Handler without logrus.
	sentry, err := logrus_sentry.NewSentryHook(value, nil)
	if err != nil {
		log.Warn("Probably Sentry host is not running", "err", err)
		return
	}

	log.Root().SetHandler(
		log.MultiHandler(
			log.Root().GetHandler(),
			LogrusHandler(sentry),
		))
}

// SetLevel sets level filter on log root handler.
// So it should be called last.
func SetLevel(l string) {
	lvl, err := log.LvlFromString(l)
	if err != nil {
		panic(err)
	}

	log.Root().SetHandler(
		log.LvlFilterHandler(
			lvl,
			log.Root().GetHandler()))
}
