package drivermodule

import (
	"io"
	"math"
	"math/big"

	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/prometheus/common/log"

	"github.com/Fantom-foundation/go-opera/gossip/blockproc"
	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/inter/drivertype"
	"github.com/Fantom-foundation/go-opera/inter/validatorpk"
	"github.com/Fantom-foundation/go-opera/opera"
	"github.com/Fantom-foundation/go-opera/opera/genesis"
	"github.com/Fantom-foundation/go-opera/opera/genesis/driver"
	"github.com/Fantom-foundation/go-opera/opera/genesis/driver/drivercall"
	"github.com/Fantom-foundation/go-opera/opera/genesis/driver/driverpos"
	"github.com/Fantom-foundation/go-opera/opera/genesis/driverauth"
	"github.com/Fantom-foundation/go-opera/opera/genesis/netinit"
	netinitcall "github.com/Fantom-foundation/go-opera/opera/genesis/netinit/netinitcalls"
	"github.com/Fantom-foundation/go-opera/opera/genesis/sfc"
)

const (
	maxAdvanceEpochs = 1 << 16
)

type DriverTxListenerModule struct{}

func NewDriverTxListenerModule() *DriverTxListenerModule {
	return &DriverTxListenerModule{}
}

func (m *DriverTxListenerModule) Start(block blockproc.BlockCtx, bs blockproc.BlockState, es blockproc.EpochState, statedb *state.StateDB) blockproc.TxListener {
	return &DriverTxListener{
		block:   block,
		es:      es,
		bs:      bs,
		statedb: statedb,
	}
}

type DriverTxListener struct {
	block   blockproc.BlockCtx
	es      blockproc.EpochState
	bs      blockproc.BlockState
	statedb *state.StateDB
}

type DriverTxTransactor struct{}

type DriverTxPreTransactor struct{}

type DriverTxGenesisTransactor struct {
	g opera.Genesis
}

func NewDriverTxTransactor() *DriverTxTransactor {
	return &DriverTxTransactor{}
}

func NewDriverTxPreTransactor() *DriverTxPreTransactor {
	return &DriverTxPreTransactor{}
}

func NewDriverTxGenesisTransactor(g opera.Genesis) *DriverTxGenesisTransactor {
	return &DriverTxGenesisTransactor{
		g: g,
	}
}

func internalTxBuilder(statedb *state.StateDB) func(calldata []byte, addr common.Address) *types.Transaction {
	nonce := uint64(math.MaxUint64)
	return func(calldata []byte, addr common.Address) *types.Transaction {
		if nonce == math.MaxUint64 {
			nonce = statedb.GetNonce(common.Address{})
		}
		tx := types.NewTransaction(nonce, addr, common.Big0, 1e10, common.Big0, calldata)
		nonce++
		return tx
	}
}

func (p *DriverTxGenesisTransactor) PopInternalTxs(_ blockproc.BlockCtx, _ blockproc.BlockState, es blockproc.EpochState, _ bool, statedb *state.StateDB) types.Transactions {
	buildTx := internalTxBuilder(statedb)
	internalTxs := make(types.Transactions, 0, 15)
	// initialization
	calldata := netinitcall.InitializeAll(es.Epoch, p.g.TotalSupply, sfc.ContractAddress, driverauth.ContractAddress, driver.ContractAddress, p.g.DriverOwner)
	internalTxs = append(internalTxs, buildTx(calldata, netinit.ContractAddress))
	// push genesis validators
	for _, v := range p.g.Validators {
		calldata := drivercall.SetGenesisValidator(v)
		internalTxs = append(internalTxs, buildTx(calldata, driver.ContractAddress))
	}
	// push genesis delegations
	p.g.Delegations.ForEach(func(addr common.Address, toValidatorID idx.ValidatorID, delegation genesis.Delegation) {
		if delegation.Stake.Sign() == 0 {
			panic(addr.String())
		}
		calldata := drivercall.SetGenesisDelegation(addr, toValidatorID, delegation)
		internalTxs = append(internalTxs, buildTx(calldata, driver.ContractAddress))
	})
	return internalTxs
}

func (p *DriverTxPreTransactor) PopInternalTxs(block blockproc.BlockCtx, bs blockproc.BlockState, es blockproc.EpochState, sealing bool, statedb *state.StateDB) types.Transactions {
	buildTx := internalTxBuilder(statedb)
	internalTxs := make(types.Transactions, 0, 8)

	// write cheaters
	for _, validatorID := range block.CBlock.Cheaters {
		valIdx := es.Validators.GetIdx(validatorID)
		if bs.ValidatorStates[valIdx].Cheater {
			continue
		}
		bs.ValidatorStates[valIdx].Cheater = true
		calldata := drivercall.DeactivateValidator(validatorID, drivertype.DoublesignBit)
		internalTxs = append(internalTxs, buildTx(calldata, driver.ContractAddress))
	}

	// push data into Driver before epoch sealing
	if sealing {
		metrics := make([]drivercall.ValidatorEpochMetric, es.Validators.Len())
		for oldValIdx := idx.Validator(0); oldValIdx < es.Validators.Len(); oldValIdx++ {
			info := bs.ValidatorStates[oldValIdx]
			missed := opera.BlocksMissed{
				BlocksNum: block.Idx - info.LastBlock,
				Period:    inter.MaxTimestamp(block.Time, info.LastMedianTime) - info.LastMedianTime,
			}
			if missed.BlocksNum <= es.Rules.Economy.BlockMissedSlack {
				missed = opera.BlocksMissed{}
			}
			metrics[oldValIdx] = drivercall.ValidatorEpochMetric{
				Missed:          missed,
				Uptime:          info.Uptime,
				OriginatedTxFee: info.Originated,
			}
		}
		calldata := drivercall.SealEpoch(metrics)
		internalTxs = append(internalTxs, buildTx(calldata, driver.ContractAddress))
	}
	return internalTxs
}

func (p *DriverTxTransactor) PopInternalTxs(_ blockproc.BlockCtx, _ blockproc.BlockState, es blockproc.EpochState, sealing bool, statedb *state.StateDB) types.Transactions {
	buildTx := internalTxBuilder(statedb)
	internalTxs := make(types.Transactions, 0, 1)
	// push data into Driver after epoch sealing
	if sealing {
		calldata := drivercall.SealEpochValidators(es.Validators.SortedIDs())
		internalTxs = append(internalTxs, buildTx(calldata, driver.ContractAddress))
	}
	return internalTxs
}

func (p *DriverTxListener) OnNewReceipt(tx *types.Transaction, r *types.Receipt, originator idx.ValidatorID) {
	if originator == 0 {
		return
	}
	originatorIdx := p.es.Validators.GetIdx(originator)

	// track originated fee
	txFee := new(big.Int).Mul(new(big.Int).SetUint64(r.GasUsed), tx.GasPrice())
	originated := p.bs.ValidatorStates[originatorIdx].Originated
	originated.Add(originated, txFee)

	// track gas power refunds
	notUsedGas := tx.Gas() - r.GasUsed
	if notUsedGas != 0 {
		p.bs.ValidatorStates[originatorIdx].DirtyGasRefund += notUsedGas
	}
}

func decodeDataBytes(l *types.Log) ([]byte, error) {
	if len(l.Data) < 32 {
		return nil, io.ErrUnexpectedEOF
	}
	start := new(big.Int).SetBytes(l.Data[24:32]).Uint64()
	if start+32 > uint64(len(l.Data)) {
		return nil, io.ErrUnexpectedEOF
	}
	size := new(big.Int).SetBytes(l.Data[start+24 : start+32]).Uint64()
	if start+32+size > uint64(len(l.Data)) {
		return nil, io.ErrUnexpectedEOF
	}
	return l.Data[start+32 : start+32+size], nil
}

func (p *DriverTxListener) OnNewLog(l *types.Log) {
	if l.Address != driver.ContractAddress {
		return
	}
	// Track validator weight changes
	if l.Topics[0] == driverpos.Topics.UpdateValidatorWeight && len(l.Topics) > 1 && len(l.Data) >= 32 {
		validatorID := idx.ValidatorID(new(big.Int).SetBytes(l.Topics[1][:]).Uint64())
		weight := new(big.Int).SetBytes(l.Data[0:32])

		if weight.Sign() == 0 {
			delete(p.bs.NextValidatorProfiles, validatorID)
		} else {
			profile := p.bs.NextValidatorProfiles[validatorID]
			profile.Weight = weight
			p.bs.NextValidatorProfiles[validatorID] = profile
		}
	}
	// Track validator pubkey changes
	if l.Topics[0] == driverpos.Topics.UpdateValidatorPubkey && len(l.Topics) > 1 {
		validatorID := idx.ValidatorID(new(big.Int).SetBytes(l.Topics[1][:]).Uint64())
		pubkey, err := decodeDataBytes(l)
		if err != nil {
			log.Warn("Malformed UpdatedValidatorPubkey Driver event")
			return
		}

		profile := p.bs.NextValidatorProfiles[validatorID]
		profile.PubKey, _ = validatorpk.FromBytes(pubkey)
		p.bs.NextValidatorProfiles[validatorID] = profile
	}
	// Add balance
	if l.Topics[0] == driverpos.Topics.IncBalance && len(l.Topics) > 1 && len(l.Data) >= 32 {
		acc := common.BytesToAddress(l.Topics[1][12:])
		value := new(big.Int).SetBytes(l.Data[0:32])

		p.statedb.AddBalance(acc, value)
	}
	// Set balance
	if l.Topics[0] == driverpos.Topics.SetBalance && len(l.Topics) > 1 && len(l.Data) >= 32 {
		acc := common.BytesToAddress(l.Topics[1][12:])
		value := new(big.Int).SetBytes(l.Data[0:32])

		p.statedb.SetBalance(acc, value)
	}
	// Subtract balance
	if l.Topics[0] == driverpos.Topics.SubBalance && len(l.Topics) > 1 && len(l.Data) >= 32 {
		acc := common.BytesToAddress(l.Topics[1][12:])
		value := new(big.Int).SetBytes(l.Data[0:32])

		if p.statedb.GetBalance(acc).Cmp(value) <= 0 {
			p.statedb.SetBalance(acc, new(big.Int))
		} else {
			p.statedb.SubBalance(acc, value)
		}
	}
	// Set code
	if l.Topics[0] == driverpos.Topics.SetCode && len(l.Topics) > 2 {
		acc := common.BytesToAddress(l.Topics[1][12:])
		from := common.BytesToAddress(l.Topics[2][12:])

		code := p.statedb.GetCode(from)
		if code == nil {
			code = []byte{}
		}
		p.statedb.SetCode(acc, code)
	}
	// Swap code
	if l.Topics[0] == driverpos.Topics.SwapCode && len(l.Topics) > 2 {
		acc0 := common.BytesToAddress(l.Topics[1][12:])
		acc1 := common.BytesToAddress(l.Topics[2][12:])

		code0 := p.statedb.GetCode(acc0)
		if code0 == nil {
			code0 = []byte{}
		}
		code1 := p.statedb.GetCode(acc1)
		if code1 == nil {
			code1 = []byte{}
		}
		p.statedb.SetCode(acc0, code1)
		p.statedb.SetCode(acc1, code0)
	}
	// Set storage
	if l.Topics[0] == driverpos.Topics.SetStorage && len(l.Topics) > 1 && len(l.Data) >= 64 {
		acc := common.BytesToAddress(l.Topics[1][12:])
		key := common.BytesToHash(l.Data[0:32])
		value := common.BytesToHash(l.Data[32:64])

		p.statedb.SetState(acc, key, value)
	}
	// Update rules
	if l.Topics[0] == driverpos.Topics.UpdateNetworkRules && len(l.Data) >= 64 {
		diff, err := decodeDataBytes(l)
		if err != nil {
			log.Warn("Malformed UpdateNetworkRules Driver event")
			return
		}

		p.bs.DirtyRules, err = opera.UpdateRules(p.bs.DirtyRules, diff)
		if err != nil {
			log.Warn("Network rules update error", "err", err)
			return
		}
	}
	// Advance epochs
	if l.Topics[0] == driverpos.Topics.UpdateNetworkRules && len(l.Data) >= 32 {
		// epochsNum < 2^24 to avoid overflow
		epochsNum := new(big.Int).SetBytes(l.Data[29:32]).Uint64()

		p.bs.AdvanceEpochs += idx.Epoch(epochsNum)
		if p.bs.AdvanceEpochs > maxAdvanceEpochs {
			p.bs.AdvanceEpochs = maxAdvanceEpochs
		}
	}
}

func (p *DriverTxListener) Update(bs blockproc.BlockState, es blockproc.EpochState) {
	p.bs, p.es = bs, es
}

func (p *DriverTxListener) Finalize() blockproc.BlockState {
	return p.bs
}
