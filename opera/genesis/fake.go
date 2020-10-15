package genesis

import (
	"crypto/ecdsa"
	"math/big"
	"math/rand"

	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/Fantom-foundation/go-opera/inter/validator"
	"github.com/Fantom-foundation/go-opera/opera/genesis/gpos"
)

// FakeKey gets n-th fake private key.
func FakeKey(n int) *ecdsa.PrivateKey {
	reader := rand.New(rand.NewSource(int64(n)))

	key, err := ecdsa.GenerateKey(crypto.S256(), reader)
	if err != nil {
		panic(err)
	}

	return key
}

// FakeValidators returns validators accounts for fakenet
func FakeValidators(count int, balance *big.Int, stake *big.Int) VAccounts {
	accs := make(Accounts, count)
	validators := make(gpos.Validators, 0, count)
	var admin common.Address

	for i := 1; i <= count; i++ {
		key := FakeKey(i)
		addr := crypto.PubkeyToAddress(key.PublicKey)
		if i == 1 {
			admin = addr
		}
		accs[addr] = Account{
			Balance:    balance,
			PrivateKey: key,
		}
		pubkeyraw := crypto.FromECDSAPub(&key.PublicKey)
		validatorID := idx.ValidatorID(i)
		validators = append(validators, gpos.Validator{
			ID:      validatorID,
			Address: addr,
			PubKey: validator.PubKey{
				Raw:  pubkeyraw,
				Type: "secp256k1",
			},
			Stake: stake,
		})
	}

	return VAccounts{Accounts: accs, Validators: validators, SfcContractAdmin: admin}
}
