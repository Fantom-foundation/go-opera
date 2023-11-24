package launcher

import (
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/ethereum/go-ethereum/params"

	"github.com/Fantom-foundation/go-opera/opera"
	"github.com/Fantom-foundation/go-opera/opera/genesis"
	"github.com/Fantom-foundation/go-opera/opera/genesisstore"
)

var (
	Bootnodes = map[string][]string{
		"main": {},
		"x1-testnet": {
			"enode://9ed669d0c35cc4eb7aba930d8c05b5ad8ff0e3318ad1c3c5f81ae6241b40698ef4146b2a9ac80343130df2e83f137f78648565fd810b090595c0b8b4b4123a48@34.211.87.93:5050",
			"enode://c03ae9cb7a3485aba9ae3945944fad9bc678bc362bf13741033511f689df7f22063b312fac900f3fcf2ef3792e6312a84204842964085a7f0b44f4c850500405@35.155.140.94:5050",
			"enode://db70d9620aeb252d18cd7309405393921da47acf2e13cd1559b190bcc5d554bf6a75c2675c2f9bb29710dcb48ae94ef9b9abe9a9baf5ab7348c35e53f3c47c8c@44.224.180.169:5050",
		},
	}

	testnetHeader = genesis.Header{
		GenesisID:   hash.HexToHash("0x4c4fdf6346c8851355eb305399d05036a0512aea15d1cb9b364a353704d5fbcb"),
		NetworkID:   opera.TestNetworkID,
		NetworkName: "x1-testnet",
	}

	AllowedOperaGenesis = []GenesisTemplate{
		{
			Name:   "x1-testnet with full MPT",
			Header: testnetHeader,
			Hashes: genesis.Hashes{
				genesisstore.EpochsSection(0): hash.HexToHash("0xe4c9d47ea5aac4beaac9655c8e63257762f1a4bc4e55028765d7b791392beba7"),
				genesisstore.BlocksSection(0): hash.HexToHash("0xc3163030010b8a02bdce6dbb8d6dacb2e9d9a136c3afe03a4700df6473e8252a"),
				genesisstore.EvmSection(0):    hash.HexToHash("0x48c9563f4c42333b1c8c0a6feccd839c5feb3c9507e334ddb088fc2ee67b4641"),
			},
		},
	}
)

func overrideParams() {
	params.MainnetBootnodes = []string{}
	params.RopstenBootnodes = []string{}
	params.RinkebyBootnodes = []string{}
	params.GoerliBootnodes = []string{}
}
