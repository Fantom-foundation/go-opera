// +build windows

package lachesis_log

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc/eventlog"
)

// SyslogHook to send logs via syslog.
type SyslogHook struct {
	Writer *eventlog.Log
}

// Creates a hook to be added to an instance of logger. This is called with
// `hook, err := NewSyslogHook("", "localhost", "MySource")`
// `if err == nil { log.Hooks.Add(hook) }`
func NewSyslogHook(network, raddr string, src string) (*SyslogHook, error) {
	// Continue if we receive "registry key already exists" or if we get
	// ERROR_ACCESS_DENIED so that we can log without administrative permissions
	// for pre-existing eventlog sources.
	if err := eventlog.InstallAsEventCreate(src, eventlog.Info|eventlog.Warning|eventlog.Error); err != nil {
		if !strings.Contains(err.Error(), "registry key already exists") && err != windows.ERROR_ACCESS_DENIED {
			return nil, err
		}
	}
	el, err := eventlog.OpenRemote(raddr, src)
	if err != nil {
		return nil, err
	}
	return &SyslogHook{el}, err
}

func (hook *SyslogHook) Fire(entry *logrus.Entry) error {
	line, err := entry.String()
	if err != nil {
		return fmt.Errorf("Unable to read entry, %v", err)
	}

	switch entry.Level {
	case logrus.PanicLevel:
		return hook.Writer.Error(1, line)
	case logrus.FatalLevel:
		return hook.Writer.Error(2, line)
	case logrus.ErrorLevel:
		return hook.Writer.Error(3, line)
	case logrus.WarnLevel:
		return hook.Writer.Warning(4, line)
	case logrus.InfoLevel:
		return hook.Writer.Info(5, line)
	case logrus.DebugLevel, logrus.TraceLevel:
		return hook.Writer.Info(6, line)
	default:
		return nil
	}
}

func (hook *SyslogHook) Levels() []logrus.Level {
	return logrus.AllLevels
}
