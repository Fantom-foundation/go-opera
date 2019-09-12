package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	cli "gopkg.in/urfave/cli.v1"

	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	"github.com/Fantom-foundation/go-lachesis/src/lachesis"
)

var FakeNetFlag = cli.StringFlag{
	Name:  "fakenet",
	Usage: "'N/X' - sets fake N-th key and genesis of X keys",
}

func setFakeNodeConfig(ctx *cli.Context, cfg *node.Config) {
	if !ctx.GlobalIsSet(FakeNetFlag.Name) {
		return
	}

	num, _, err := parseFakeGen(ctx.GlobalString(FakeNetFlag.Name))
	if err != nil {
		log.Crit("invalid flag", "flag", FakeNetFlag.Name, "err", err)
	}

	cfg.P2P.PrivateKey = crypto.FakeKey(num)
}

func setFakeNetConfig(ctx *cli.Context, cfg *lachesis.Config) {
	if !ctx.GlobalIsSet(FakeNetFlag.Name) {
		return
	}

	_, total, err := parseFakeGen(ctx.GlobalString(FakeNetFlag.Name))
	if err != nil {
		log.Crit("invalid flag", "flag", FakeNetFlag.Name, "err", err)
	}

	*cfg = lachesis.FakeNetConfig(total)
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
