package main

import (
	"crypto/ecdsa"
	"fmt"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	cli "gopkg.in/urfave/cli.v1"

	"github.com/Fantom-foundation/go-lachesis/src/crypto"
)

var FakeNetFlag = cli.StringFlag{
	Name:  "fakenet",
	Usage: "'N/X' - sets fake N-th key and genesis of X keys",
}

func addFakeAccount(ctx *cli.Context, stack *node.Node) {
	if !ctx.GlobalIsSet(FakeNetFlag.Name) {
		return
	}

	const pswd = "fakepassword"

	kss := stack.AccountManager().Backends(keystore.KeyStoreType)
	if len(kss) < 1 {
		log.Warn("no keystore for fake accounts")
		return
	}
	ks := kss[0].(*keystore.KeyStore)

	coinbase, key := getFakeCoinbase(ctx)

	_, err := ks.ImportECDSA(key, pswd)
	if err != nil && err.Error() != "account already exists" {
		log.Crit("failed to import fake key", "err", err)
	}

	err = ks.Unlock(coinbase, pswd)
	if err != nil {
		log.Crit("failed to unlock fake key", "err", err)
	}
	log.Info("Unlocked fake coinbase", "address", coinbase.Address.Hex())
}

func getFakeCoinbase(ctx *cli.Context) (accounts.Account, *ecdsa.PrivateKey) {
	num, _, err := parseFakeGen(ctx.GlobalString(FakeNetFlag.Name))
	if err != nil {
		log.Crit("invalid flag", "flag", FakeNetFlag.Name, "err", err)
	}

	key := crypto.FakeKey(num)

	return accounts.Account{
		Address: crypto.PubkeyToAddress(key.PublicKey),
	}, key

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
