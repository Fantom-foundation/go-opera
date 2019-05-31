package lachesis

import (
	"testing"
	"time"
)

func TestService(t *testing.T) {
	l := NewForTests(nil, "server.fake", nil, nil)
	l.serviceStart()
	defer l.serviceStop()

	<-time.After(time.Second)
}
