package logger

import (
	"sync"

	"github.com/evalphobia/logrus_sentry"
	"github.com/Gurpartap/logrus-stack"
	"github.com/sirupsen/logrus"
)

var (
	log  *logrus.Logger
	once sync.Once
)

// SetDSN for sentry logger
func SetDSN(value string) {
	// If DSN is empty, we don't create new hook.
	// Otherwise we'll the same error message for each new log.
	if value == "" {
		log.Warn("Sentry client DSN is empty")
		return
	}

	hook, err := logrus_sentry.NewSentryHook(value, []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
		logrus.DebugLevel,
		logrus.TraceLevel,
	})

	if err != nil {
		log.Warn("Probably Sentry host is not running.", err)
		return
	}

	log.Hooks.Add(hook)
}

// Init sets or creates logger instance.
func Init(custom *logrus.Logger) {
	once.Do(func() {
		if custom != nil {
			log = custom
		} else {
			log = logrus.StandardLogger()
			log.SetLevel(logrus.DebugLevel)

			log.Hooks.Add(logrus_stack.StandardHook())
		}
	})
}

// Get returns logger instance.
func Get() *logrus.Logger {
	if log == nil {
		Init(nil)
	}
	return log
}

// GetLevel returns logrus.Level string representation.
func GetLevel(l string) logrus.Level {
	switch l {
	case "debug":
		return logrus.DebugLevel
	case "info":
		return logrus.InfoLevel
	case "warn":
		return logrus.WarnLevel
	case "error":
		return logrus.ErrorLevel
	case "fatal":
		return logrus.FatalLevel
	case "panic":
		return logrus.PanicLevel
	default:
		return logrus.DebugLevel
	}
}

// SetLevel sets logrus.Level.
func SetLevel(l string) {
	log.SetLevel(GetLevel(l))
}
