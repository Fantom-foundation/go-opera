package makefakegenesis

import (
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/log"
	"math/big"
	"time"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/inter/pos"
	"github.com/Fantom-foundation/lachesis-base/kvdb/memorydb"
	"github.com/Fantom-foundation/lachesis-base/lachesis"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/Fantom-foundation/go-opera/evmcore"
	"github.com/Fantom-foundation/go-opera/integration/makegenesis"
	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/inter/drivertype"
	"github.com/Fantom-foundation/go-opera/inter/iblockproc"
	"github.com/Fantom-foundation/go-opera/inter/ier"
	"github.com/Fantom-foundation/go-opera/inter/validatorpk"
	"github.com/Fantom-foundation/go-opera/opera"
	"github.com/Fantom-foundation/go-opera/opera/contracts/driver"
	"github.com/Fantom-foundation/go-opera/opera/contracts/driver/drivercall"
	"github.com/Fantom-foundation/go-opera/opera/contracts/driverauth"
	"github.com/Fantom-foundation/go-opera/opera/contracts/evmwriter"
	"github.com/Fantom-foundation/go-opera/opera/contracts/netinit"
	netinitcall "github.com/Fantom-foundation/go-opera/opera/contracts/netinit/netinitcalls"
	"github.com/Fantom-foundation/go-opera/opera/contracts/sfc"
	"github.com/Fantom-foundation/go-opera/opera/contracts/sfclib"
	"github.com/Fantom-foundation/go-opera/opera/genesis"
	"github.com/Fantom-foundation/go-opera/opera/genesis/gpos"
	"github.com/Fantom-foundation/go-opera/opera/genesisstore"
)

var (
	FakeGenesisTime = inter.Timestamp(1608600000 * time.Second)
)

// FakeKey gets n-th fake private key.
func FakeKey(n idx.ValidatorID) *ecdsa.PrivateKey {
	return evmcore.FakeKey(uint32(n))
}

func FakeGenesisStore(num idx.Validator, balance, stake *big.Int) *genesisstore.Store {
	return FakeGenesisStoreWithRules(num, balance, stake, opera.FakeNetRules())
}

func FakeGenesisStoreWithRules(num idx.Validator, balance, stake *big.Int, rules opera.Rules) *genesisstore.Store {
	return FakeGenesisStoreWithRulesAndStart(num, balance, stake, rules, 2, 1)
}

func FakeGenesisStoreWithRulesAndStart(num idx.Validator, balance, stake *big.Int, rules opera.Rules, epoch idx.Epoch, block idx.Block) *genesisstore.Store {
	builder := makegenesis.NewGenesisBuilder(memorydb.NewProducer(""))

	validators := GetFakeValidators(num)

	// add balances to validators
	var delegations []drivercall.Delegation
	for _, val := range validators {
		log.Info("Validator", "address", val.Address, "pk", val.PubKey, "id", val.ID)

		builder.AddBalance(val.Address, balance)
		delegations = append(delegations, drivercall.Delegation{
			Address:            val.Address,
			ValidatorID:        val.ID,
			Stake:              stake,
			LockedStake:        new(big.Int),
			LockupFromEpoch:    0,
			LockupEndTime:      0,
			LockupDuration:     0,
			EarlyUnlockPenalty: new(big.Int),
			Rewards:            new(big.Int),
		})
	}

	// deploy essential contracts
	// pre deploy NetworkInitializer
	builder.SetCode(netinit.ContractAddress, netinit.GetContractBin())
	// pre deploy NodeDriver
	builder.SetCode(driver.ContractAddress, driver.GetContractBin())
	// pre deploy NodeDriverAuth
	builder.SetCode(driverauth.ContractAddress, driverauth.GetContractBin())
	// pre deploy SFC
	builder.SetCode(sfc.ContractAddress, sfc.GetContractBin())
	// pre deploy SFCLib
	builder.SetCode(sfclib.ContractAddress, sfclib.GetContractBin())
	// set non-zero code for pre-compiled contracts
	builder.SetCode(evmwriter.ContractAddress, []byte{0})

	builder.SetCurrentEpoch(ier.LlrIdxFullEpochRecord{
		LlrFullEpochRecord: ier.LlrFullEpochRecord{
			BlockState: iblockproc.BlockState{
				LastBlock: iblockproc.BlockCtx{
					Idx:     block - 1,
					Time:    FakeGenesisTime,
					Atropos: hash.Event{},
				},
				FinalizedStateRoot:    hash.Hash{},
				EpochGas:              0,
				EpochCheaters:         lachesis.Cheaters{},
				CheatersWritten:       0,
				ValidatorStates:       make([]iblockproc.ValidatorBlockState, 0),
				NextValidatorProfiles: make(map[idx.ValidatorID]drivertype.Validator),
				DirtyRules:            nil,
				AdvanceEpochs:         0,
			},
			EpochState: iblockproc.EpochState{
				Epoch:             epoch - 1,
				EpochStart:        FakeGenesisTime,
				PrevEpochStart:    FakeGenesisTime - 1,
				EpochStateRoot:    hash.Zero,
				Validators:        pos.NewBuilder().Build(),
				ValidatorStates:   make([]iblockproc.ValidatorEpochState, 0),
				ValidatorProfiles: make(map[idx.ValidatorID]drivertype.Validator),
				Rules:             rules,
			},
		},
		Idx: epoch - 1,
	})

	var owner common.Address
	if num != 0 {
		owner = validators[0].Address
	}

	blockProc := makegenesis.DefaultBlockProc()
	genesisTxs := GetGenesisTxs(epoch-2, validators, builder.TotalSupply(), delegations, owner)
	err := builder.ExecuteGenesisTxs(blockProc, genesisTxs)
	if err != nil {
		panic(err)
	}

	return builder.Build(genesis.Header{
		GenesisID:   builder.CurrentHash(),
		NetworkID:   rules.NetworkID,
		NetworkName: rules.Name,
	})
}

func txBuilder() func(calldata []byte, addr common.Address) *types.Transaction {
	nonce := uint64(0)
	return func(calldata []byte, addr common.Address) *types.Transaction {
		tx := types.NewTransaction(nonce, addr, common.Big0, 1e10, common.Big0, calldata)
		nonce++
		return tx
	}
}

func GetGenesisTxs(sealedEpoch idx.Epoch, validators gpos.Validators, totalSupply *big.Int, delegations []drivercall.Delegation, driverOwner common.Address) types.Transactions {
	buildTx := txBuilder()
	internalTxs := make(types.Transactions, 0, 15)
	// initialization
	calldata := netinitcall.InitializeAll(sealedEpoch, totalSupply, sfc.ContractAddress, sfclib.ContractAddress, driverauth.ContractAddress, driver.ContractAddress, evmwriter.ContractAddress, driverOwner)
	internalTxs = append(internalTxs, buildTx(calldata, netinit.ContractAddress))
	// push genesis validators
	for _, v := range validators {
		calldata := drivercall.SetGenesisValidator(v)
		internalTxs = append(internalTxs, buildTx(calldata, driver.ContractAddress))
	}
	// push genesis delegations
	for _, delegation := range delegations {
		calldata := drivercall.SetGenesisDelegation(delegation)
		internalTxs = append(internalTxs, buildTx(calldata, driver.ContractAddress))
	}
	return internalTxs
}

func GetFakeValidators(num idx.Validator) gpos.Validators {
	validators := make(gpos.Validators, 0, num)

	for i := idx.ValidatorID(1); i <= idx.ValidatorID(num); i++ {
		key := FakeKey(i)
		addr := crypto.PubkeyToAddress(key.PublicKey)
		pubkeyraw := crypto.FromECDSAPub(&key.PublicKey)

		publicKey := validatorpk.PubKey{
			Raw:  pubkeyraw,
			Type: validatorpk.Types.Secp256k1,
		}

		log.Info("publicKey", "key", publicKey, "addr", addr)

		validators = append(validators, gpos.Validator{
			ID:               i,
			Address:          addr,
			PubKey:           publicKey,
			CreationTime:     FakeGenesisTime,
			CreationEpoch:    0,
			DeactivatedTime:  0,
			DeactivatedEpoch: 0,
			Status:           0,
		})
	}

	return validators
}
