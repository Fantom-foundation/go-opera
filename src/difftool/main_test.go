package difftool

import (
	"strings"
	"testing"

	"github.com/andrecronje/lachesis/src/common"
)

func TestComparing(t *testing.T) {
	logger := common.NewTestLogger(t)

	nodes := NewNodeList(4, logger)

	stopTxStream := nodes.StartRandTxStream()

	nodes.WaitForBlock(2)

	stopTxStream()

	output := Compare(nodes.Nodes()...)

	t.Log(strings.Join(output, "\n"))
}
