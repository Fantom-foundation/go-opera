package logger

import (
	"testing"
)

// SetTestMode sets test mode.
func SetTestMode(t testing.TB) {
	log.Out = &testLoggerAdapter{
		t: t,
	}
}

// This can be used as the destination for a logger and it'll
// map them into calls to testing.T.Log, so that you only see
// the logging for failed tests.
type testLoggerAdapter struct {
	t testing.TB
}

// Write implements io.Writer.
func (a *testLoggerAdapter) Write(d []byte) (int, error) {
	if d[len(d)-1] == '\n' {
		d = d[:len(d)-1]
	}
	a.t.Log(string(d))
	return len(d), nil
}
