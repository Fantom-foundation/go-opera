package posnode

import (
	"github.com/sirupsen/logrus"
)

type logger struct {
	log *logrus.Entry
}

func newLogger(node string) logger {
	return logger{
		log: log.WithField("node", node),
	}
}

/*
 * global vars:
 */

var (
	log *logrus.Logger
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
	log = custom
}
