package lachesis

import (
	"testing"
	"time"

	"github.com/Fantom-foundation/go-lachesis/src/logger"
)

func TestService(t *testing.T) {
	logger.SetTestMode(t)

	l := NewForTests(nil, "server.fake", nil, nil)
	l.serviceStart()
	defer l.serviceStop()

	<-time.After(time.Second)
}
