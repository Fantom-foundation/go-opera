package launcher

import (
	"crypto/ecdsa"
	"fmt"
	"strconv"
	"strings"

	"github.com/Fantom-foundation/go-opera/inter/validatorpk"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	cli "gopkg.in/urfave/cli.v1"

	"github.com/Fantom-foundation/go-opera/integration/makefakegenesis"
)

// FakeNetFlag enables special testnet, where validators are automatically created
var FakeNetFlag = cli.StringFlag{
	Name:  "fakenet",
	Usage: "'n/N' - sets coinbase as fake n-th key from genesis of [1..N] validators. Use n=0 for non-validator node",
}

func getFakeValidatorID(ctx *cli.Context) idx.ValidatorID {
	id, _, err := parseFakeGen(ctx.GlobalString(FakeNetFlag.Name))
	if err != nil {
		log.Crit("Invalid flag", "flag", FakeNetFlag.Name, "err", err)
	}
	return id
}

func getFakeValidatorKey(ctx *cli.Context) *ecdsa.PrivateKey {
	id, _, err := parseFakeGen(ctx.GlobalString(FakeNetFlag.Name))
	if err != nil {
		log.Crit("Invalid flag", "flag", FakeNetFlag.Name, "err", err)
	}
	return makefakegenesis.FakeKey(id)
}

func getFakeValidatorCount(ctx *cli.Context) idx.Validator {
	_, num, err := parseFakeGen(ctx.GlobalString(FakeNetFlag.Name))
	if err != nil {
		log.Crit("Invalid flag", "flag", FakeNetFlag.Name, "err", err)
		return 0
	}
	return num
}

func parseFakeGen(s string) (id idx.ValidatorID, num idx.Validator, err error) {
	parts := strings.SplitN(s, "/", 2)
	if len(parts) != 2 {
		err = fmt.Errorf("use %%d/%%d format")
		return
	}

	var u32 uint64
	u32, err = strconv.ParseUint(parts[0], 10, 32)
	if err != nil {
		return
	}
	id = idx.ValidatorID(u32)

	u32, err = strconv.ParseUint(parts[1], 10, 32)
	if err != nil {
		return
	}
	num = idx.Validator(u32)

	return
}

func fakeValidatorPubKey(id idx.ValidatorID) validatorpk.PubKey {
	return validatorpk.PubKey{
		Raw:  crypto.FromECDSAPub(&makefakegenesis.FakeKey(id).PublicKey),
		Type: validatorpk.Types.Secp256k1,
	}
}
