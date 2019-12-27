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
	genesisTestTime = inter.Timestamp(1565000000 * time.Second)
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
		Time:  genesisTestTime,
	}
	g = preDeploySfc(g, sfc.GetTestContractBinV1())
	return g
}

// MainGenesis returns builtin genesis keys of mainnet.
func MainGenesis() Genesis {
	g := Genesis{
		Time: genesisTestTime,
		Alloc: VAccounts{
			Accounts: Accounts{
				// TODO: fill with official keys and balances before release!
				common.HexToAddress("a123456789123456789123456789012345678901"): Account{Balance: utils.ToFtm(1e10)},
				common.HexToAddress("a123456789123456789123456789012345678902"): Account{Balance: utils.ToFtm(1e10)},
				common.HexToAddress("a123456789123456789123456789012345678903"): Account{Balance: utils.ToFtm(1e10)},
			},
			Validators: pos.GValidators{
				pos.GenesisValidator{
					ID:      1,
					Address: common.HexToAddress("a123456789123456789123456789012345678901"),
					Stake:   utils.ToFtm(3175000),
				},
				pos.GenesisValidator{
					ID:      2,
					Address: common.HexToAddress("a123456789123456789123456789012345678902"),
					Stake:   utils.ToFtm(3175000),
				},
			},
			SfcContractAdmin: common.HexToAddress("a123456789123456789123456789012345678904"),
		},
	}
	g = preDeploySfc(g, sfc.GetMainContractBinV1())
	return g
}

// TestGenesis returns builtin genesis keys of testnet.
func TestGenesis() Genesis {
	g := Genesis{
		Time: genesisTestTime,
		Alloc: VAccounts{
			Accounts: Accounts{
				// TODO: fill with official keys and balances before release!
				common.HexToAddress("b123456789123456789123456789012345678901"): Account{Balance: utils.ToFtm(1e10)},
				common.HexToAddress("b123456789123456789123456789012345678902"): Account{Balance: utils.ToFtm(1e10)},
				common.HexToAddress("b123456789123456789123456789012345678903"): Account{Balance: utils.ToFtm(1e10)},
			},
			Validators: pos.GValidators{
				pos.GenesisValidator{
					ID:      1,
					Address: common.HexToAddress("b123456789123456789123456789012345678901"),
					Stake:   utils.ToFtm(3175000),
				},
				pos.GenesisValidator{
					ID:      2,
					Address: common.HexToAddress("b123456789123456789123456789012345678902"),
					Stake:   utils.ToFtm(3175000),
				},
			},
			SfcContractAdmin: common.HexToAddress("b123456789123456789123456789012345678904"),
		},
	}
	g = preDeploySfc(g, sfc.GetTestContractBinV1())
	return g
}
