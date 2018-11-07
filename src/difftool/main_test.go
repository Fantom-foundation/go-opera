package difftool

import (
	"testing"
	"time"

	"github.com/andrecronje/lachesis/src/common"
)

func TestComparing(t *testing.T) {
	logger := common.NewTestLogger(t)

	nodes := NewNodeList(4, logger)

	nodes.PushRandTxs(100)

	<-time.After(10 * time.Second)

	output := Compare(nodes.Nodes()...)

	t.Log(output)
}
