package main

import (
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/common"
	cli "gopkg.in/urfave/cli.v1"

	"github.com/Fantom-foundation/go-lachesis/crypto"
	"github.com/Fantom-foundation/go-lachesis/gossip"
)

var coinbaseFlag = cli.StringFlag{
	Name:  "coinbase",
	Usage: "Public address for block mining rewards",
	Value: "0",
}

// setCoinbase retrieves the etherbase either from the directly specified
// command line flags or from the keystore if CLI indexed.
func setCoinbase(ctx *cli.Context, ks *keystore.KeyStore, cfg *gossip.EmitterConfig) {
	// Extract the current coinbase, new flag overriding legacy one
	var coinbase string
	switch {
	case ctx.GlobalIsSet(coinbaseFlag.Name):
		coinbase = ctx.GlobalString(coinbaseFlag.Name)
		if coinbase == "no" {
			coinbase = common.Address{}.String()
		}
	case ctx.GlobalIsSet(utils.MinerEtherbaseFlag.Name):
		coinbase = ctx.GlobalString(utils.MinerEtherbaseFlag.Name)
	case ctx.GlobalIsSet(utils.MinerLegacyEtherbaseFlag.Name):
		coinbase = ctx.GlobalString(utils.MinerLegacyEtherbaseFlag.Name)
	case ctx.GlobalIsSet(FakeNetFlag.Name):
		key := getFakeCoinbase(ctx)
		coinbase = crypto.PubkeyToAddress(key.PublicKey).Hex()
	}

	// Convert the etherbase into an address and configure it
	if coinbase != "" {
		if ks != nil {
			account, err := utils.MakeAddress(ks, coinbase)
			if err != nil {
				utils.Fatalf("Invalid miner etherbase: %v", err)
			}
			cfg.Coinbase = account.Address
		} else {
			utils.Fatalf("No etherbase configured")
		}
	}
}
