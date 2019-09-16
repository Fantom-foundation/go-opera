package main

import (
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/cmd/utils"
	cli "gopkg.in/urfave/cli.v1"

	"github.com/Fantom-foundation/go-lachesis/src/gossip"
)

// setEtherbase retrieves the etherbase either from the directly specified
// command line flags or from the keystore if CLI indexed.
func setEtherbase(ctx *cli.Context, ks *keystore.KeyStore, cfg *gossip.EmitterConfig) {
	// Extract the current etherbase, new flag overriding legacy one
	var etherbase string
	if ctx.GlobalIsSet(utils.MinerLegacyEtherbaseFlag.Name) {
		etherbase = ctx.GlobalString(utils.MinerLegacyEtherbaseFlag.Name)
	}
	if ctx.GlobalIsSet(utils.MinerEtherbaseFlag.Name) {
		etherbase = ctx.GlobalString(utils.MinerEtherbaseFlag.Name)
	}
	// Convert the etherbase into an address and configure it
	if etherbase != "" {
		if ks != nil {
			account, err := utils.MakeAddress(ks, etherbase)
			if err != nil {
				utils.Fatalf("Invalid miner etherbase: %v", err)
			}
			cfg.Emitbase = account.Address
		} else {
			utils.Fatalf("No etherbase configured")
		}
	}
}
