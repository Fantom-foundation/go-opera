package genesis

import (
	"math/big"

	"github.com/Fantom-foundation/go-lachesis/crypto"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/inter/pos"
)

// FakeAccounts returns accounts and validators for fakenet
func FakeAccounts(from, count int, balance *big.Int, stake *big.Int) VAccounts {
	accs := make(Accounts, count)
	validators := make(pos.GValidators, 0, count)

	for i := from; i < from+count; i++ {
		key := crypto.FakeKey(i)
		addr := crypto.PubkeyToAddress(key.PublicKey)
		accs[addr] = Account{
			Balance:    balance,
			PrivateKey: key,
		}
		stakerID := idx.StakerID(i + 1)
		validators = append(validators, pos.GenesisValidator{
			ID:      stakerID,
			Address: addr,
			Stake:   stake,
		})
	}

	return VAccounts{Accounts: accs, Validators: validators, SfcContractAdmin: validators[0].Address}
}
