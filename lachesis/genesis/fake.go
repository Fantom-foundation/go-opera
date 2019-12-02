package genesis

import (
	"github.com/Fantom-foundation/go-lachesis/crypto"
	"github.com/Fantom-foundation/go-lachesis/inter/pos"
)

func FakeAccounts(from, count int, stake pos.Stake) Accounts {
	accs := make(Accounts, count)

	for i := from; i < from+count; i++ {
		key := crypto.FakeKey(i)
		addr := crypto.PubkeyToAddress(key.PublicKey)
		accs[addr] = Account{
			Balance:    pos.StakeToBalance(stake),
			PrivateKey: key,
		}
	}

	return accs
}
