package main

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	cli "gopkg.in/urfave/cli.v1"

	"github.com/Fantom-foundation/go-lachesis/crypto"
	"github.com/Fantom-foundation/go-lachesis/integration"
	"github.com/Fantom-foundation/go-lachesis/lachesis/genesis"
	"github.com/Fantom-foundation/go-lachesis/utils"
)

// FakeNetFlag enables special testnet, where validators are automatically created
var FakeNetFlag = cli.StringFlag{
	Name:  "fakenet",
	Usage: "'n/N[,non-validators]' - sets coinbase as fake n-th key from genesis of N validators. Non-validators is a count or json-file.",
}

func addFakeAccount(ctx *cli.Context, stack *node.Node) {
	if !ctx.GlobalIsSet(FakeNetFlag.Name) {
		return
	}

	key := getFakeValidator(ctx)
	if key == nil {
		return
	}

	coinbase := integration.SetAccountKey(stack.AccountManager(), key, "fakepassword")
	log.Info("Unlocked fake validator", "address", coinbase.Address.Hex())
}

func getFakeValidator(ctx *cli.Context) *ecdsa.PrivateKey {
	num, _, err := parseFakeGen(ctx.GlobalString(FakeNetFlag.Name))
	if err != nil {
		log.Crit("Invalid flag", "flag", FakeNetFlag.Name, "err", err)
	}

	if num == 0 {
		return nil
	}

	return crypto.FakeKey(num)
}

func parseFakeGen(s string) (num int, vaccs genesis.VAccounts, err error) {
	var i64 uint64

	parts := strings.SplitN(s, "/", 2)
	if len(parts) != 2 {
		err = fmt.Errorf("use %%d/%%d format")
		return
	}

	i64, err = strconv.ParseUint(parts[0], 10, 32)
	if err != nil {
		return
	}
	num = int(i64)

	parts = strings.SplitN(parts[1], ",", 2)

	i64, err = strconv.ParseUint(parts[0], 10, 32)
	validatorsNum := int(i64)

	if validatorsNum < 0 || num > validatorsNum {
		err = fmt.Errorf("key-num should be in range from 1 to validators (<key-num>/<validators>), or should be zero for non-validator node")
		return
	}

	vaccs = genesis.FakeValidators(validatorsNum, utils.ToFtm(1e10), utils.ToFtm(3175000))

	if len(parts) < 2 {
		return
	}
	var others genesis.Accounts
	i64, err = strconv.ParseUint(parts[1], 10, 32)
	if err != nil {
		others, err = readAccounts(parts[1])
	} else {
		others = genAccounts(validatorsNum, int(i64), big.NewInt(1e18), big.NewInt(0))
	}
	vaccs.Accounts.Add(others)

	return
}

func readAccounts(filename string) (accs genesis.Accounts, err error) {
	var f *os.File
	f, err = os.Open(filename)
	if err != nil {
		return
	}

	accs = genesis.Accounts{}
	err = json.NewDecoder(f).Decode(&accs)
	return
}

func genAccounts(offset, count int, balance *big.Int, stake *big.Int) genesis.Accounts {
	accs := make(genesis.Accounts, count)

	for i := offset; i < offset+count; i++ {
		key := crypto.FakeKey(i)
		addr := crypto.PubkeyToAddress(key.PublicKey)
		accs[addr] = genesis.Account{
			Balance:    balance,
			PrivateKey: key,
		}
	}

	return accs
}
