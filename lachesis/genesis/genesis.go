package genesis

import (
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
	Alloc      Accounts
	Validators pos.Validators
	Time       inter.Timestamp
	ExtraData  []byte
}

func preDeploySfc(g Genesis) Genesis {
	g.Alloc[sfc.ContractAddress] = Account{
		Code:    sfc.GetContractBinV1(),
		Storage: sfc.AssembleStorage(g.Validators, g.Time, nil),
		Balance: pos.StakeToBalance(g.Validators.TotalStake()),
	}
	return g
}

// FakeGenesis generates fake genesis with n-nodes.
func FakeGenesis(accs VAccounts) Genesis {
	g := Genesis{
		Alloc:      accs.Accounts,
		Validators: accs.Validators,
		Time:       genesisTestTime,
	}
	g = preDeploySfc(g)
	return g
}

// MainGenesis returns builtin genesis keys of mainnet.
func MainGenesis() Genesis {
	validators := pos.NewValidators()
	validators.Set(common.HexToAddress("a123456789123456789123456789012345678901"), 1000000000000)
	validators.Set(common.HexToAddress("a123456789123456789123456789012345678902"), 1000000000000)

	g := Genesis{
		Time: genesisTestTime,
		Alloc: Accounts{
			// TODO: fill with official keys and balances before release!
			common.HexToAddress("a123456789123456789123456789012345678901"): Account{Balance: pos.StakeToBalance(50000000000000)},
			common.HexToAddress("a123456789123456789123456789012345678902"): Account{Balance: pos.StakeToBalance(50000000000000)},
		},
		Validators: *validators,
	}
	g = preDeploySfc(g)
	return g
}

// TestGenesis returns builtin genesis keys of testnet.
func TestGenesis() Genesis {
	validators := pos.NewValidators()
	validators.Set(common.HexToAddress("b123456789123456789123456789012345678901"), 1000000000000)
	validators.Set(common.HexToAddress("b123456789123456789123456789012345678902"), 1000000000000)

	g := Genesis{
		Time: genesisTestTime,
		Alloc: Accounts{
			// TODO: fill with official keys and balances before release!
			common.HexToAddress("b123456789123456789123456789012345678901"): Account{Balance: pos.StakeToBalance(50000000000000)},
			common.HexToAddress("b123456789123456789123456789012345678902"): Account{Balance: pos.StakeToBalance(50000000000000)},
		},
		Validators: *validators,
	}
	g = preDeploySfc(g)
	return g
}
