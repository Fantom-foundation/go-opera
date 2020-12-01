package sfcmodule

import (
	"math"
	"math/big"

	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/prometheus/common/log"

	"github.com/Fantom-foundation/go-opera/gossip/blockproc"
	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/inter/sfctype"
	"github.com/Fantom-foundation/go-opera/opera"
	"github.com/Fantom-foundation/go-opera/opera/genesis"
	"github.com/Fantom-foundation/go-opera/opera/genesis/sfc"
	"github.com/Fantom-foundation/go-opera/opera/genesis/sfc/sfccall"
	"github.com/Fantom-foundation/go-opera/opera/genesis/sfc/sfcpos"
)

type SfcTxListenerModule struct {
	net opera.Rules
}

func NewSfcTxListenerModule(net opera.Rules) *SfcTxListenerModule {
	return &SfcTxListenerModule{
		net: net,
	}
}

func (m *SfcTxListenerModule) Start(block blockproc.BlockCtx, bs blockproc.BlockState, es blockproc.EpochState, statedb *state.StateDB) blockproc.TxListener {
	return &SfcTxListener{
		block:   block,
		es:      es,
		bs:      bs,
		statedb: statedb,
		net:     m.net,
	}
}

type SfcTxListener struct {
	block   blockproc.BlockCtx
	es      blockproc.EpochState
	bs      blockproc.BlockState
	statedb *state.StateDB

	net opera.Rules
}

type SfcTxTransactor struct {
	net opera.Rules
}

type SfcTxPreTransactor struct {
	net opera.Rules
}

type SfcTxGenesisTransactor struct {
	g opera.Genesis
}

func NewSfcTxTransactor(net opera.Rules) *SfcTxTransactor {
	return &SfcTxTransactor{
		net: net,
	}
}

func NewSfcTxPreTransactor(net opera.Rules) *SfcTxPreTransactor {
	return &SfcTxPreTransactor{
		net: net,
	}
}

func NewSfcTxGenesisTransactor(g opera.Genesis) *SfcTxGenesisTransactor {
	return &SfcTxGenesisTransactor{
		g: g,
	}
}

func internalTxBuilder(statedb *state.StateDB) func(calldata []byte) *types.Transaction {
	nonce := uint64(math.MaxUint64)
	return func(calldata []byte) *types.Transaction {
		if nonce == math.MaxUint64 {
			nonce = statedb.GetNonce(common.Address{})
		}
		tx := types.NewTransaction(nonce, sfc.ContractAddress, common.Big0, 1e10, common.Big0, calldata)
		nonce++
		return tx
	}
}

func (p *SfcTxGenesisTransactor) PopInternalTxs(_ blockproc.BlockCtx, _ blockproc.BlockState, _ blockproc.EpochState, _ bool, statedb *state.StateDB) types.Transactions {
	buildTx := internalTxBuilder(statedb)
	internalTxs := make(types.Transactions, 0, 15)
	// push genesis validators
	for _, v := range p.g.State.Validators {
		calldata := sfccall.SetGenesisValidator(v)
		internalTxs = append(internalTxs, buildTx(calldata))
	}
	// push genesis delegations
	p.g.State.Delegations.ForEach(func(addr common.Address, toValidatorID idx.ValidatorID, delegation genesis.Delegation) {
		if delegation.Stake.Sign() == 0 {
			panic(addr.String())
		}
		calldata := sfccall.SetGenesisDelegation(addr, toValidatorID, delegation)
		internalTxs = append(internalTxs, buildTx(calldata))
	})
	// finish initialization
	calldata := sfccall.Initialize(0)
	internalTxs = append(internalTxs, buildTx(calldata))
	return internalTxs
}

func (p *SfcTxPreTransactor) PopInternalTxs(block blockproc.BlockCtx, bs blockproc.BlockState, es blockproc.EpochState, sealing bool, statedb *state.StateDB) types.Transactions {
	buildTx := internalTxBuilder(statedb)
	internalTxs := make(types.Transactions, 0, 8)

	// write cheaters
	for _, validatorID := range block.CBlock.Cheaters {
		valIdx := es.Validators.GetIdx(validatorID)
		if bs.ValidatorStates[valIdx].Cheater {
			continue
		}
		bs.ValidatorStates[valIdx].Cheater = true
		calldata := sfccall.DeactivateValidator(validatorID, sfctype.DoublesignBit)
		internalTxs = append(internalTxs, buildTx(calldata))
	}

	// push data into SFC before epoch sealing
	if sealing {
		metrics := make([]sfccall.ValidatorEpochMetric, es.Validators.Len())
		for oldValIdx := 0; oldValIdx < es.Validators.Len(); oldValIdx++ {
			info := bs.ValidatorStates[oldValIdx]
			missed := opera.BlocksMissed{
				BlocksNum: block.Idx - info.LastBlock,
				Period:    inter.MaxTimestamp(block.Time, info.LastMedianTime) - info.LastMedianTime,
			}
			metrics[oldValIdx] = sfccall.ValidatorEpochMetric{
				Missed:          missed,
				Uptime:          info.Uptime,
				OriginatedTxFee: info.Originated,
			}
		}
		calldata := sfccall.SealEpoch(metrics)
		internalTxs = append(internalTxs, buildTx(calldata))
	}
	return internalTxs
}

func (p *SfcTxTransactor) PopInternalTxs(_ blockproc.BlockCtx, _ blockproc.BlockState, es blockproc.EpochState, sealing bool, statedb *state.StateDB) types.Transactions {
	buildTx := internalTxBuilder(statedb)
	internalTxs := make(types.Transactions, 0, 1)
	// push data into SFC after epoch sealing
	if sealing {
		calldata := sfccall.SealEpochValidators(es.Validators.SortedIDs())
		internalTxs = append(internalTxs, buildTx(calldata))
	}
	return internalTxs
}

func (p *SfcTxListener) OnNewReceipt(tx *types.Transaction, r *types.Receipt, originator idx.ValidatorID) {
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

func (p *SfcTxListener) OnNewLog(l *types.Log) {
	if l.Address != sfc.ContractAddress {
		return
	}
	// Track validator weight changes
	if l.Topics[0] == sfcpos.Topics.UpdatedValidatorWeight && len(l.Topics) > 1 && len(l.Data) >= 32 {
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
	if l.Topics[0] == sfcpos.Topics.UpdatedValidatorPubkey && len(l.Topics) > 1 && len(l.Data) > 64 {
		validatorID := idx.ValidatorID(new(big.Int).SetBytes(l.Topics[1][:]).Uint64())
		start := new(big.Int).SetBytes(l.Data[24:32]).Uint64()
		if start > uint64(len(l.Data)-32) {
			log.Warn("Malformed UpdatedValidatorPubkey SFC event")
			return
		}
		size := new(big.Int).SetBytes(l.Data[start+24 : start+32]).Uint64()
		if start+32+size > uint64(len(l.Data)) {
			log.Warn("Malformed UpdatedValidatorPubkey SFC event")
			return
		}
		pubkey := l.Data[start+32 : start+32+size]

		profile := p.bs.NextValidatorProfiles[validatorID]
		profile.Pubkey.Type = "secp256k1"
		profile.Pubkey.Raw = pubkey
		p.bs.NextValidatorProfiles[validatorID] = profile
	}
	// Add balance
	if l.Topics[0] == sfcpos.Topics.IncBalance && len(l.Topics) > 1 && len(l.Data) >= 32 {
		acc := common.BytesToAddress(l.Topics[1][12:])
		value := new(big.Int).SetBytes(l.Data[0:32])

		p.statedb.AddBalance(acc, value)
	}
}

func (p *SfcTxListener) Update(bs blockproc.BlockState, es blockproc.EpochState) {
	p.bs, p.es = bs, es
}

func (p *SfcTxListener) Finalize() blockproc.BlockState {
	return p.bs
}
