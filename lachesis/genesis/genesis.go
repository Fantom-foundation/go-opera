package genesis

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/pos"
	"github.com/Fantom-foundation/go-lachesis/lachesis/genesis/proxy"
	"github.com/Fantom-foundation/go-lachesis/lachesis/genesis/sfc"
	"github.com/Fantom-foundation/go-lachesis/utils"
)

var (
	genesisTime = inter.Timestamp(1577419000 * time.Second)
)

type Genesis struct {
	Alloc     VAccounts
	Time      inter.Timestamp
	ExtraData []byte
}

func preDeploySfc(g Genesis, implCode []byte) Genesis {
	// pre deploy SFC impl
	g.Alloc.Accounts[sfc.ContractAddressV1] = Account{
		Code:    implCode, // impl account has only code, balance and storage is in proxy account
		Balance: big.NewInt(0),
	}
	// pre deploy SFC proxy
	storage := sfc.AssembleStorage(g.Alloc.Validators, g.Time, g.Alloc.SfcContractAdmin, nil)
	storage = proxy.AssembleStorage(g.Alloc.SfcContractAdmin, sfc.ContractAddressV1, storage) // Add storage of proxy
	g.Alloc.Accounts[sfc.ContractAddress] = Account{
		Code:    proxy.GetContractBin(),
		Storage: storage,
		Balance: g.Alloc.Validators.TotalStake(),
	}
	return g
}

// FakeGenesis generates fake genesis with n-nodes.
func FakeGenesis(accs VAccounts) Genesis {
	g := Genesis{
		Alloc: accs,
		Time:  genesisTime,
	}
	g = preDeploySfc(g, sfc.GetTestContractBinV1())
	return g
}

// MainGenesis returns builtin genesis keys of mainnet.
func MainGenesis() Genesis {
	g := Genesis{
		Time: genesisTime,
		Alloc: VAccounts{
			Accounts: Accounts{
				common.HexToAddress("0xd6A37423Be930019b8CFeA57BE049329f3119a3D"): Account{Balance: utils.ToFtm(2000000100)},
				common.HexToAddress("0x541E408443A592C38e01Bed0cB31f9De8c1322d0"): Account{Balance: utils.ToFtm(100)},
				common.HexToAddress("0x35701189D211215Cb38393f407B4767886DeB03A"): Account{Balance: utils.ToFtm(100)},
				common.HexToAddress("0xEfC9200cD50ae935DA5d79D122660DDB53620E74"): Account{Balance: utils.ToFtm(100)},
				common.HexToAddress("0xe9f77B989E3a73ed24D0467fAA9300Ab94477915"): Account{Balance: utils.ToFtm(100)},
				common.HexToAddress("0x7f9d1dbaf84d827b0840e38f555a490969978d20"): Account{Balance: utils.ToFtm(100)},
				common.HexToAddress("0xfd09f0296af88ac777c137ecd92d85583a9b9e4a"): Account{Balance: utils.ToFtm(100)},
				common.HexToAddress("0x520D07E2F0F3f60b7510e65a291862976a9547c6"): Account{Balance: utils.ToFtm(100)},
				common.HexToAddress("0x93194b30FA85D1927852769bAB1822dD2B6818e1"): Account{Balance: utils.ToFtm(100)},
				common.HexToAddress("0xc73b84eac512a04f139D4a200aAe0FF24c4A5cBC"): Account{Balance: utils.ToFtm(100)},
				common.HexToAddress("0xd160D9B59508e4636eEc3E0a7f734268D1cE1047"): Account{Balance: utils.ToFtm(100)},
				common.HexToAddress("0x03894D86CC6e1E41e8bBa17469204D304C4e92b8"): Account{Balance: utils.ToFtm(100)},
				common.HexToAddress("0x51AF0532E6BD89695266E714a1B09fD1Ac297DC7"): Account{Balance: utils.ToFtm(100)},
			},
			Validators: pos.GValidators{
				pos.GenesisValidator{
					ID:      1,
					Address: common.HexToAddress("0x541E408443A592C38e01Bed0cB31f9De8c1322d0"),
					Stake:   utils.ToFtm(10000000),
				},
				pos.GenesisValidator{
					ID:      2,
					Address: common.HexToAddress("0x35701189D211215Cb38393f407B4767886DeB03A"),
					Stake:   utils.ToFtm(10000000),
				},
				pos.GenesisValidator{
					ID:      3,
					Address: common.HexToAddress("0xEfC9200cD50ae935DA5d79D122660DDB53620E74"),
					Stake:   utils.ToFtm(10000000),
				},
				pos.GenesisValidator{
					ID:      4,
					Address: common.HexToAddress("0xe9f77B989E3a73ed24D0467fAA9300Ab94477915"),
					Stake:   utils.ToFtm(10000000),
				},
				pos.GenesisValidator{
					ID:      5,
					Address: common.HexToAddress("0x7f9d1dbaf84d827b0840e38f555a490969978d20"),
					Stake:   utils.ToFtm(10000000),
				},
				pos.GenesisValidator{
					ID:      6,
					Address: common.HexToAddress("0xfd09f0296af88ac777c137ecd92d85583a9b9e4a"),
					Stake:   utils.ToFtm(10000000),
				},
				pos.GenesisValidator{
					ID:      7,
					Address: common.HexToAddress("0x520D07E2F0F3f60b7510e65a291862976a9547c6"),
					Stake:   utils.ToFtm(10000000),
				},
				pos.GenesisValidator{
					ID:      8,
					Address: common.HexToAddress("0x93194b30FA85D1927852769bAB1822dD2B6818e1"),
					Stake:   utils.ToFtm(10000000),
				},
				pos.GenesisValidator{
					ID:      9,
					Address: common.HexToAddress("0xc73b84eac512a04f139D4a200aAe0FF24c4A5cBC"),
					Stake:   utils.ToFtm(10000000),
				},
				pos.GenesisValidator{
					ID:      10,
					Address: common.HexToAddress("0xd160D9B59508e4636eEc3E0a7f734268D1cE1047"),
					Stake:   utils.ToFtm(10000000),
				},
				pos.GenesisValidator{
					ID:      11,
					Address: common.HexToAddress("0x03894D86CC6e1E41e8bBa17469204D304C4e92b8"),
					Stake:   utils.ToFtm(10000000),
				},
				pos.GenesisValidator{
					ID:      12,
					Address: common.HexToAddress("0x51AF0532E6BD89695266E714a1B09fD1Ac297DC7"),
					Stake:   utils.ToFtm(10000000),
				},
			},
			SfcContractAdmin: common.HexToAddress("0xd6A37423Be930019b8CFeA57BE049329f3119a3D"),
		},
	}
	g = preDeploySfc(g, sfc.GetMainContractBinV1())
	return g
}

// TestGenesis returns builtin genesis keys of testnet.
func TestGenesis() Genesis {
	g := Genesis{
		Time: genesisTime,
		Alloc: VAccounts{
			Accounts: Accounts{
				common.HexToAddress("0xe003e080e8d61207a0a9890c3663b4cd7fb766b8"): Account{Balance: utils.ToFtm(2000000100)},
				common.HexToAddress("0xcc8b10332478e26f676bccfc73f8c687e3ad1d04"): Account{Balance: utils.ToFtm(400)},
				common.HexToAddress("0x30e3b5cc7e8fb98a22e688dfb20b327be8a9fe30"): Account{Balance: utils.ToFtm(400)},
				common.HexToAddress("0x567b6f3d4ba1f55652cf90df6db90ad6d8f9abc1"): Account{Balance: utils.ToFtm(400)},
			},
			Validators: pos.GValidators{
				pos.GenesisValidator{
					ID:      1,
					Address: common.HexToAddress("0xcc8b10332478e26f676bccfc73f8c687e3ad1d04"),
					Stake:   utils.ToFtm(40000000),
				},
				pos.GenesisValidator{
					ID:      2,
					Address: common.HexToAddress("0x30e3b5cc7e8fb98a22e688dfb20b327be8a9fe30"),
					Stake:   utils.ToFtm(40000000),
				},
				pos.GenesisValidator{
					ID:      3,
					Address: common.HexToAddress("0x567b6f3d4ba1f55652cf90df6db90ad6d8f9abc1"),
					Stake:   utils.ToFtm(40000000),
				},
			},
			SfcContractAdmin: common.HexToAddress("0xe003e080e8d61207a0a9890c3663b4cd7fb766b8"),
		},
	}
	g = preDeploySfc(g, sfc.GetTestContractBinV1())
	return g
}
