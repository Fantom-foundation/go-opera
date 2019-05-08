package dummy

import (
	"testing"

	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/proxy"
)

func TestProxyAppImplementation(t *testing.T) {
	logger := common.NewTestLogger(t)

	state := interface{}(
		NewState(logger))

	_, ok := state.(proxy.App)
	if !ok {
		t.Fatal("State does not implement proxy.App interface!")
	}
}
