package posposet_test

import (
	"testing"
	"time"

	"github.com/Fantom-foundation/go-lachesis/src/posposet"
)

func TestStartStop(t *testing.T) {
	p := posposet.New(posposet.GenerateKey())
	p.Stop()
	p.Start()
	p.Start()
	<-time.After(100 * time.Microsecond)
	p.Stop()
	p.Stop()
}
