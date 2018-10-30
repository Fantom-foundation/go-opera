package dummy

import (
	"testing"

	"github.com/andrecronje/lachesis/src/common"
	"github.com/andrecronje/lachesis/src/proxy"
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
