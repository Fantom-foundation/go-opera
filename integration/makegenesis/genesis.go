package makegenesis

import (
	"crypto/ecdsa"
	"math/big"
	"math/rand"
	"time"

	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/kvdb/leveldb"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/syndtr/goleveldb/leveldb/opt"

	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/inter/validator"
	"github.com/Fantom-foundation/go-opera/opera"
	"github.com/Fantom-foundation/go-opera/opera/genesis"
	"github.com/Fantom-foundation/go-opera/opera/genesis/gpos"
	"github.com/Fantom-foundation/go-opera/opera/genesis/proxy"
	"github.com/Fantom-foundation/go-opera/opera/genesis/proxy/proxypos"
	"github.com/Fantom-foundation/go-opera/opera/genesis/sfc"
	"github.com/Fantom-foundation/go-opera/opera/genesisstore"
)

var (
	FakeGenesisTime = inter.Timestamp(1577419000 * time.Second)
)

func OpenGenesis(path string) (*genesisstore.Store, error) {
	db, err := leveldb.New(path, opt.MiB, 0, nil, nil)
	if err != nil {
		return nil, err
	}
	return genesisstore.NewStore(db), nil
}

// FakeKey gets n-th fake private key.
func FakeKey(n int) *ecdsa.PrivateKey {
	reader := rand.New(rand.NewSource(int64(n)))

	key, err := ecdsa.GenerateKey(crypto.S256(), reader)
	if err != nil {
		panic(err)
	}

	return key
}

func FakeGenesisStore(num int, balance, stake *big.Int) *genesisstore.Store {
	genStore := genesisstore.NewMemStore()
	genStore.SetRules(opera.FakeNetRules())

	validators := GetFakeValidators(num)

	for _, val := range validators {
		genStore.SetEvmAccount(val.Address, genesis.Account{
			Code:    []byte{},
			Balance: balance,
			Nonce:   0,
		})
		genStore.SetDelegation(val.Address, val.ID, genesis.Delegation{
			Stake:   stake,
			Rewards: new(big.Int),
		})
	}

	var admin common.Address
	if num != 0 {
		admin = validators[0].Address
	}

	genStore.SetMetadata(genesisstore.Metadata{
		Validators:    validators,
		FirstEpoch:    1,
		Time:          FakeGenesisTime,
		PrevEpochTime: FakeGenesisTime - inter.Timestamp(time.Hour),
		ExtraData:     []byte("fake"),
	})
	preDeploySfc(admin, sfc.ContractAddress, sfc.ContractAddressV1, proxy.GetContractBin(), sfc.GetContractBin(), genStore)

	return genStore
}

func GetFakeValidators(num int) gpos.Validators {
	validators := make(gpos.Validators, 0, num)

	for i := 1; i <= num; i++ {
		key := FakeKey(i)
		addr := crypto.PubkeyToAddress(key.PublicKey)
		pubkeyraw := crypto.FromECDSAPub(&key.PublicKey)
		validatorID := idx.ValidatorID(i)
		validators = append(validators, gpos.Validator{
			ID:      validatorID,
			Address: addr,
			PubKey: validator.PubKey{
				Raw:  pubkeyraw,
				Type: "secp256k1",
			},
			CreationTime:     FakeGenesisTime,
			CreationEpoch:    0,
			DeactivatedTime:  0,
			DeactivatedEpoch: 0,
			Status:           0,
		})
	}

	return validators
}

func preDeploySfc(admin, proxyAddr, implAddr common.Address, proxyCode, implCode []byte, genStore *genesisstore.Store) {
	// pre deploy SFC impl
	genStore.SetEvmAccount(implAddr, genesis.Account{
		Code:    implCode,
		Balance: new(big.Int),
		Nonce:   0,
	})
	// pre deploy SFC proxy
	genStore.SetEvmAccount(proxyAddr, genesis.Account{
		Code:    proxyCode,
		Balance: new(big.Int),
		Nonce:   0,
	})
	genStore.SetEvmState(proxyAddr, proxypos.Admin(), admin.Hash())
	genStore.SetEvmState(proxyAddr, proxypos.Implementation(), implAddr.Hash())
}
