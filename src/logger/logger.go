package logger

import (
	"github.com/sirupsen/logrus"
)

/*
 * global vars:
 */

var (
	Log *logrus.Logger
)

func init() {
	defaults := logrus.StandardLogger()
	defaults.SetLevel(logrus.DebugLevel)
	SetLogger(defaults)
}

// SetLogger sets logger for whole package.
func SetLogger(custom *logrus.Logger) {
	if custom == nil {
		panic("Nil-logger set")
	}
	Log = custom
}

// GetLevel return logrus.Level by string.
// Example: "debug" -> logrus.DebugLevel and etc.
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
