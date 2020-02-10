package main

import (
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/cmd/utils"
	cli "gopkg.in/urfave/cli.v1"

	"github.com/Fantom-foundation/go-lachesis/crypto"
	"github.com/Fantom-foundation/go-lachesis/gossip"
)

var validatorFlag = cli.StringFlag{
	Name:  "validator",
	Usage: "Address of a validator to create events from",
	Value: "no",
}

// setValidator retrieves the validator address either from the directly specified
// command line flags or from the keystore if CLI indexed.
func setValidator(ctx *cli.Context, ks *keystore.KeyStore, cfg *gossip.EmitterConfig) {
	// Extract the current validator address, new flag overriding legacy one
	var validator string
	switch {
	case ctx.GlobalIsSet(validatorFlag.Name):
		validator = ctx.GlobalString(validatorFlag.Name)
		if validator == "no" || validator == "0" {
			validator = ""
		}
	case ctx.GlobalIsSet(utils.MinerEtherbaseFlag.Name):
		validator = ctx.GlobalString(utils.MinerEtherbaseFlag.Name)
	case ctx.GlobalIsSet(utils.MinerLegacyEtherbaseFlag.Name):
		validator = ctx.GlobalString(utils.MinerLegacyEtherbaseFlag.Name)
	case ctx.GlobalIsSet(FakeNetFlag.Name):
		key := getFakeValidator(ctx)
		if key != nil {
			validator = crypto.PubkeyToAddress(key.PublicKey).Hex()
		}
	}

	// Convert the validator into an address and configure it
	if validator == "" {
		return
	}

	if ks != nil {
		account, err := utils.MakeAddress(ks, validator)
		if err != nil {
			utils.Fatalf("Invalid miner etherbase: %v", err)
		}
		cfg.Validator = account.Address
	} else {
		utils.Fatalf("No etherbase configured")
	}

}
