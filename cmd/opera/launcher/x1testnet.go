package launcher

import (
	"crypto/ecdsa"
	"github.com/Fantom-foundation/go-opera/integration/makefakegenesis"
	cli "gopkg.in/urfave/cli.v1"
)

// X1TestnetFlag enables special testnet, where validators are automatically created
var X1TestnetFlag = cli.BoolFlag{
	Name:  "x1-testnet",
	Usage: "Generates X1 testnet genesis and starts the chain",
}

func getX1ValidatorKey(ctx *cli.Context) *ecdsa.PrivateKey {
	return makefakegenesis.FakeKey(1)
}
