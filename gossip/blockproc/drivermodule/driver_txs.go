package drivermodule

import (
	"io"
	"math"
	"math/big"

	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"

	"github.com/Fantom-foundation/go-opera/evmcore"
	"github.com/Fantom-foundation/go-opera/gossip/blockproc"
	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/inter/drivertype"
	"github.com/Fantom-foundation/go-opera/inter/iblockproc"
	"github.com/Fantom-foundation/go-opera/inter/validatorpk"
	"github.com/Fantom-foundation/go-opera/opera"
	"github.com/Fantom-foundation/go-opera/opera/contracts/driver"
	"github.com/Fantom-foundation/go-opera/opera/contracts/driver/drivercall"
	"github.com/Fantom-foundation/go-opera/opera/contracts/driver/driverpos"
)

const (
	maxAdvanceEpochs = 1 << 16
)

type DriverTxListenerModule struct{}

func NewDriverTxListenerModule() *DriverTxListenerModule {
	return &DriverTxListenerModule{}
}

func (m *DriverTxListenerModule) Start(block iblockproc.BlockCtx, bs iblockproc.BlockState, es iblockproc.EpochState, statedb evmcore.StateDB) blockproc.TxListener {
	return &DriverTxListener{
		block:   block,
		es:      es,
		bs:      bs,
		statedb: statedb,
	}
}

type DriverTxListener struct {
	block   iblockproc.BlockCtx
	es      iblockproc.EpochState
	bs      iblockproc.BlockState
	statedb evmcore.StateDB
}

type DriverTxTransactor struct{}

type DriverTxPreTransactor struct{}

func NewDriverTxTransactor() *DriverTxTransactor {
	return &DriverTxTransactor{}
}

func NewDriverTxPreTransactor() *DriverTxPreTransactor {
	return &DriverTxPreTransactor{}
}

func InternalTxBuilder(statedb evmcore.StateDB) func(calldata []byte, addr common.Address) *types.Transaction {
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

func maxBlockIdx(a, b idx.Block) idx.Block {
	if a > b {
		return a
	}
	return b
}

func (p *DriverTxPreTransactor) PopInternalTxs(block iblockproc.BlockCtx, bs iblockproc.BlockState, es iblockproc.EpochState, sealing bool, statedb evmcore.StateDB) types.Transactions {
	buildTx := InternalTxBuilder(statedb)
	internalTxs := make(types.Transactions, 0, 8)

	// write cheaters
	for _, validatorID := range bs.EpochCheaters[bs.CheatersWritten:] {
		calldata := drivercall.DeactivateValidator(validatorID, drivertype.DoublesignBit)
		internalTxs = append(internalTxs, buildTx(calldata, driver.ContractAddress))
	}

	// push data into Driver before epoch sealing
	if sealing {
		metrics := make([]drivercall.ValidatorEpochMetric, es.Validators.Len())
		for oldValIdx := idx.Validator(0); oldValIdx < es.Validators.Len(); oldValIdx++ {
			info := bs.ValidatorStates[oldValIdx]
			// forgive downtime if below BlockMissedSlack
			missed := opera.BlocksMissed{
				BlocksNum: maxBlockIdx(block.Idx, info.LastBlock) - info.LastBlock,
				Period:    inter.MaxTimestamp(block.Time, info.LastOnlineTime) - info.LastOnlineTime,
			}
			uptime := info.Uptime
			if missed.BlocksNum <= es.Rules.Economy.BlockMissedSlack {
				missed = opera.BlocksMissed{}
				prevOnlineTime := inter.MaxTimestamp(info.LastOnlineTime, es.EpochStart)
				uptime += inter.MaxTimestamp(block.Time, prevOnlineTime) - prevOnlineTime
			}
			metrics[oldValIdx] = drivercall.ValidatorEpochMetric{
				Missed:          missed,
				Uptime:          uptime,
				OriginatedTxFee: info.Originated,
			}
		}
		calldata := drivercall.SealEpoch(metrics)
		internalTxs = append(internalTxs, buildTx(calldata, driver.ContractAddress))
	}
	return internalTxs
}

func (p *DriverTxTransactor) PopInternalTxs(_ iblockproc.BlockCtx, _ iblockproc.BlockState, es iblockproc.EpochState, sealing bool, statedb evmcore.StateDB) types.Transactions {
	buildTx := InternalTxBuilder(statedb)
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
			profile, ok := p.bs.NextValidatorProfiles[validatorID]
			if !ok {
				profile.PubKey = validatorpk.PubKey{
					Type: 0,
					Raw:  []byte{},
				}
			}
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

		profile, ok := p.bs.NextValidatorProfiles[validatorID]
		if !ok {
			log.Warn("Unexpected UpdatedValidatorPubkey Driver event")
			return
		}
		profile.PubKey, _ = validatorpk.FromBytes(pubkey)
		p.bs.NextValidatorProfiles[validatorID] = profile
	}
	// Update rules
	if l.Topics[0] == driverpos.Topics.UpdateNetworkRules && len(l.Data) >= 64 {
		diff, err := decodeDataBytes(l)
		if err != nil {
			log.Warn("Malformed UpdateNetworkRules Driver event")
			return
		}

		last := &p.es.Rules
		if p.bs.DirtyRules != nil {
			last = p.bs.DirtyRules
		}
		updated, err := opera.UpdateRules(*last, diff)
		if err != nil {
			log.Warn("Network rules update error", "err", err)
			return
		}
		p.bs.DirtyRules = &updated
	}
	// Advance epochs
	if l.Topics[0] == driverpos.Topics.AdvanceEpochs && len(l.Data) >= 32 {
		// epochsNum < 2^24 to avoid overflow
		epochsNum := new(big.Int).SetBytes(l.Data[29:32]).Uint64()

		p.bs.AdvanceEpochs += idx.Epoch(epochsNum)
		if p.bs.AdvanceEpochs > maxAdvanceEpochs {
			p.bs.AdvanceEpochs = maxAdvanceEpochs
		}
	}
}

func (p *DriverTxListener) Update(bs iblockproc.BlockState, es iblockproc.EpochState) {
	p.bs, p.es = bs, es
}

func (p *DriverTxListener) Finalize() iblockproc.BlockState {
	return p.bs
}
