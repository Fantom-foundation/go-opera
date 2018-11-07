package difftool

import (
	"fmt"

	"github.com/andrecronje/lachesis/src/node"
)

func Compare(nodes ...*node.Node) (output []string) {
	for i := len(nodes) - 1; i > 0; i-- {
		n0 := nodes[i]
		for _, n1 := range nodes[:i] {
			diff := compare(n0, n1)
			output = append(output, diff)
		}
	}
	return
}

func compare(n0, n1 *node.Node) string {
	return fmt.Sprintf("compare %d vs %d", n0.ID(), n1.ID())
}
