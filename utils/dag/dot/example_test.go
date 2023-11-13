package dot_test

import (
	"fmt"

	"github.com/Fantom-foundation/go-opera/utils/dag/dot"
)

func ExampleNewGraph() {
	g := dot.NewGraph("G")
	g.Set("label", "Example graph")
	n1, n2 := dot.NewNode("Node1"), dot.NewNode("Node2")

	n1.Set("color", "sienna")

	g.AddNode(n1)
	g.AddNode(n2)

	e := dot.NewEdge(n1, n2)
	e.Set("dir", "both")
	g.AddEdge(e)

	fmt.Println(g)
	// Output:
	// digraph G {
	// graph [
	//   label="Example graph";
	// ];
	// Node1 [color=sienna];
	// Node2;
	// Node1 -> Node2  [ dir=both ]
	// }
	//
}
