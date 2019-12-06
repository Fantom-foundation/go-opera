package genesis

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/pos"
	"github.com/Fantom-foundation/go-lachesis/lachesis/genesis/sfc"
)

var (
	genesisTestTime = inter.Timestamp(1565000000 * time.Second)
)

type Genesis struct {
	Alloc     VAccounts
	Time      inter.Timestamp
	ExtraData []byte
}

func preDeploySfc(g Genesis) Genesis {
	g.Alloc.Accounts[sfc.ContractAddress] = Account{
		Code:    sfc.GetContractBinV1(),
		Storage: sfc.AssembleStorage(g.Alloc.GValidators, g.Time, nil),
		Balance: pos.StakeToBalance(g.Alloc.GValidators.Validators().TotalStake()),
	}
	return g
}

// FakeGenesis generates fake genesis with n-nodes.
func FakeGenesis(accs VAccounts) Genesis {
	g := Genesis{
		Alloc: accs,
		Time:  genesisTestTime,
	}
	g = preDeploySfc(g)
	return g
}

// MainGenesis returns builtin genesis keys of mainnet.
func MainGenesis() Genesis {
	g := Genesis{
		Time: genesisTestTime,
		Alloc: VAccounts{
			Accounts: Accounts{
				// TODO: fill with official keys and balances before release!
				common.HexToAddress("a123456789123456789123456789012345678901"): Account{Balance: big.NewInt(1e18)},
				common.HexToAddress("a123456789123456789123456789012345678902"): Account{Balance: big.NewInt(1e18)},
				common.HexToAddress("a123456789123456789123456789012345678903"): Account{Balance: big.NewInt(1e18)},
			},
			GValidators: pos.GValidators{
				1: pos.GenesisValidator{
					ID:      1,
					Address: common.HexToAddress("a123456789123456789123456789012345678901"),
					Stake:   10000000,
				},
				2: pos.GenesisValidator{
					ID:      1,
					Address: common.HexToAddress("a123456789123456789123456789012345678902"),
					Stake:   10000000,
				},
			},
		},
	}
	g = preDeploySfc(g)
	return g
}

// TestGenesis returns builtin genesis keys of testnet.
func TestGenesis() Genesis {
	g := Genesis{
		Time: genesisTestTime,
		Alloc: VAccounts{
			Accounts: Accounts{
				// TODO: fill with official keys and balances before release!
				common.HexToAddress("b123456789123456789123456789012345678901"): Account{Balance: big.NewInt(1e18)},
				common.HexToAddress("b123456789123456789123456789012345678902"): Account{Balance: big.NewInt(1e18)},
				common.HexToAddress("b123456789123456789123456789012345678903"): Account{Balance: big.NewInt(1e18)},
			},
			GValidators: pos.GValidators{
				1: pos.GenesisValidator{
					ID:      1,
					Address: common.HexToAddress("b123456789123456789123456789012345678901"),
					Stake:   10000000,
				},
				2: pos.GenesisValidator{
					ID:      1,
					Address: common.HexToAddress("b123456789123456789123456789012345678902"),
					Stake:   10000000,
				},
			},
		},
	}
	g = preDeploySfc(g)
	return g
}
