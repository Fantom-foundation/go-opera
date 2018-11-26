package difftool

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/Fantom-foundation/go-lachesis/src/node"
)

// Example illustrates nodes comparing
func Example() {
	logger := logrus.New()
	logger.Level = logrus.FatalLevel

	nodes := node.NewNodeList(3, logger)

	stop := nodes.StartRandTxStream()
	nodes.WaitForBlock(5)
	stop()

	diffResult := Compare(nodes.Values()...)

	if !diffResult.IsEmpty() {
		logger.Fatal("\n" + diffResult.ToString())
	}
	fmt.Println("all good")
	// Output:
	// all good
}
