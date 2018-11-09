package difftool

import (
	"strings"
	"testing"

	"github.com/andrecronje/lachesis/src/common"
)

func TestComparing(t *testing.T) {
	logger := common.NewTestLogger(t)

	nodes := NewNodeList(2, logger)

	stop := nodes.StartRandTxStream()
	nodes.WaitForBlock(3)
	stop()

	diffs := Compare(nodes.Nodes()...)

	var output []string
	for _, diff := range diffs {
		if !diff.IsEmpty() {
			output = append(output, diff.ToString())
			output = append(output, diff.Descr)
		}
	}
	t.Log("\n" + strings.Join(output, "\n"))
}
