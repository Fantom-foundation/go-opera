package dummy

import (
	"testing"

	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/proxy"
)

func TestProxyHandlerImplementation(t *testing.T) {
	logger := common.NewTestLogger(t)

	state := interface{}(
		NewState(logger))

	_, ok := state.(proxy.ProxyHandler)
	if !ok {
		t.Fatal("State does not implement ProxyHandler interface!")
	}
}
