package posnode

import (
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/Fantom-foundation/go-lachesis/src/common"
)

var (
	log         *logrus.Logger
	nodeLoggers map[common.Address]*logrus.Entry
	syncLoggers sync.RWMutex
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

	syncLoggers.Lock()
	nodeLoggers = make(map[common.Address]*logrus.Entry)
	syncLoggers.Unlock()
}

// GetLogger returns logger for node.
// TODO: use common.NodeNameDict instead of name after PR #161
func GetLogger(node common.Address, name string) *logrus.Entry {
	syncLoggers.RLock()
	l := nodeLoggers[node]
	syncLoggers.RUnlock()
	if l != nil {
		return l
	}

	if name == "" {
		name = node.String()
	}
	l = log.WithField("node", name)
	syncLoggers.Lock()
	nodeLoggers[node] = l
	syncLoggers.Unlock()
	return l
}
