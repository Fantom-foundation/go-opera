package logger

import (
	"sync"

	"github.com/sirupsen/logrus"
)

var (
	log  *logrus.Logger
	once sync.Once
)

// Init sets or creates logger instance.
func Init(custom *logrus.Logger) {
	once.Do(func() {
		if custom != nil {
			log = custom
		} else {
			log = logrus.StandardLogger()
			log.SetLevel(logrus.DebugLevel)
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
