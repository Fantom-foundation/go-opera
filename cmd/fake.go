package main

import (
	"crypto/ecdsa"
	"fmt"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	cli "gopkg.in/urfave/cli.v1"

	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	"github.com/Fantom-foundation/go-lachesis/src/integration"
)

var FakeNetFlag = cli.StringFlag{
	Name:  "fakenet",
	Usage: "'N/X' - sets fake N-th key and genesis of X keys",
}

func addFakeAccount(ctx *cli.Context, stack *node.Node) {
	if !ctx.GlobalIsSet(FakeNetFlag.Name) {
		return
	}

	key := getFakeCoinbase(ctx)
	coinbase := integration.SetAccountKey(stack.AccountManager(), key, "fakepassword")
	log.Info("Unlocked fake coinbase", "address", coinbase.Address.Hex())
}

func getFakeCoinbase(ctx *cli.Context) *ecdsa.PrivateKey {
	num, _, err := parseFakeGen(ctx.GlobalString(FakeNetFlag.Name))
	if err != nil {
		log.Crit("Invalid flag", "flag", FakeNetFlag.Name, "err", err)
	}

	return crypto.FakeKey(num)
}

func parseFakeGen(s string) (num, total int, err error) {
	parts := strings.Split(s, "/")
	if len(parts) != 2 {
		err = fmt.Errorf("use %%d/%%d format")
		return
	}

	num64, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return
	}
	num = int(num64) - 1

	total64, err := strconv.ParseUint(parts[1], 10, 64)
	total = int(total64)

	if num64 < 1 || num64 > total64 {
		err = fmt.Errorf("key-num should be in range from 1 to total : <key-num>/<total>")
	}

	return
}
