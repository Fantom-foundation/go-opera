package posposet

import (
	"github.com/sirupsen/logrus"
)

var (
	// EventNameDict is an optional dictionary to make events human readable in log.
	EventNameDict = make(map[EventHash]string)
	// logger
	log *logrus.Logger
)

func init() {
	log = logrus.StandardLogger()
	log.SetLevel(logrus.DebugLevel)
}

func SetLogger(custom *logrus.Logger) {
	if custom == nil {
		panic("Nil-logger set")
	}
	log = custom
}
