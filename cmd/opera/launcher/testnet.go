package launcher

import (
	cli "gopkg.in/urfave/cli.v1"
)

// TestnetFlag enables special testnet, where validators are automatically created
var TestnetFlag = cli.BoolFlag{
	Name:  "testnet",
	Usage: "Generates X1 testnet genesis and starts the chain",
}
