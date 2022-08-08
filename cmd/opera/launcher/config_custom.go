package launcher

import (
	"fmt"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/naoina/toml"
	"github.com/naoina/toml/ast"
)

// asDefault is slice with one empty element
// which indicates that network default bootnodes should be substituted
var asDefault = []*enode.Node{{}}

func needDefaultBootnodes(nn []*enode.Node) bool {
	return len(nn) == len(asDefault) && nn[0] == asDefault[0]
}

func isBootstrapNodesDefault(root *ast.Table) (
	bootstrapNodes bool,
	bootstrapNodesV5 bool,
) {
	table := root
	for _, path := range []string{"Node", "P2P"} {
		val, ok := table.Fields[path]
		if !ok {
			return
		}
		table = val.(*ast.Table)
	}

	emptyNode := fmt.Sprintf("\"%s\"", asDefault[0])

	var res = map[string]bool{
		"BootstrapNodes":   false,
		"BootstrapNodesV5": false,
	}
	for name := range res {
		if val, ok := table.Fields[name]; ok {
			kv := val.(*ast.KeyValue)
			arr := kv.Value.(*ast.Array)
			if len(arr.Value) == len(asDefault) && arr.Value[0].Source() == emptyNode {
				res[name] = true
				delete(table.Fields, name)
			}
		}
	}
	bootstrapNodes = res["BootstrapNodes"]
	bootstrapNodesV5 = res["BootstrapNodesV5"]

	return
}

// UnmarshalTOML implements toml.Unmarshaler.
func (c *config) UnmarshalTOML(input []byte) error {
	ast, err := toml.Parse(input)
	if err != nil {
		return err
	}

	defaultBootstrapNodes, defaultBootstrapNodesV5 := isBootstrapNodesDefault(ast)

	type rawCfg config
	var raw = rawCfg(*c)
	err = toml.UnmarshalTable(ast, &raw)
	if err != nil {
		return err
	}
	*c = config(raw)

	if defaultBootstrapNodes {
		c.Node.P2P.BootstrapNodes = asDefault
	}
	if defaultBootstrapNodesV5 {
		c.Node.P2P.BootstrapNodesV5 = asDefault
	}

	return nil
}
