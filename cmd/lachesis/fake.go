package main

import (
	"crypto/ecdsa"
	"fmt"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	cli "gopkg.in/urfave/cli.v1"

	"github.com/Fantom-foundation/go-lachesis/crypto"
	"github.com/Fantom-foundation/go-lachesis/integration"
)

// FakeNetFlag enables special testnet, where validators are automatically created
var FakeNetFlag = cli.StringFlag{
	Name:  "fakenet",
	Usage: "'N/X[,x]' - sets fake N-th key and genesis of X keys and x non-validators",
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
	num, _, _, err := parseFakeGen(ctx.GlobalString(FakeNetFlag.Name))
	if err != nil {
		log.Crit("Invalid flag", "flag", FakeNetFlag.Name, "err", err)
	}

	return crypto.FakeKey(num)
}

func parseFakeGen(s string) (num, validators, others int, err error) {
	var i64 uint64

	parts := strings.Split(s, "/")
	if len(parts) != 2 {
		err = fmt.Errorf("use %%d/%%d format")
		return
	}

	i64, err = strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return
	}
	num = int(i64) - 1

	parts = strings.Split(parts[1], ",")

	i64, err = strconv.ParseUint(parts[0], 10, 64)
	validators = int(i64)

	if validators < 1 || num >= validators {
		err = fmt.Errorf("key-num should be in range from 1 to validators : <key-num>/<validators>")
	}

	if len(parts) > 1 {
		i64, err = strconv.ParseUint(parts[1], 10, 64)
		if err != nil {
			return
		}
		others = int(i64)
	}

	return
}
