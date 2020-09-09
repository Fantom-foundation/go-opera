package basiccheck

import (
	"errors"
	"math"

	base "github.com/Fantom-foundation/lachesis-base/eventcheck/basiccheck"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/Fantom-foundation/go-opera/evmcore"
	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/opera"
	"github.com/Fantom-foundation/go-opera/opera/params"
)

var (
	ErrTooManyParents = errors.New("event has too many parents")
	ErrZeroTime       = errors.New("event has zero timestamp")
	ErrNegativeValue  = errors.New("negative value")
	ErrTooBigGasUsed  = errors.New("event uses too much gas power")
	ErrWrongGasUsed   = errors.New("event has incorrect gas power")
	ErrIntrinsicGas   = errors.New("intrinsic gas too low")
	ErrUnderpriced    = errors.New("event transaction underpriced")
)

type Checker struct {
	base   *base.Checker
	config *opera.DagConfig
}

// New validator which performs checks which don't require anything except event
func New(config *opera.DagConfig) *Checker {
	return &Checker{
		config: config,
		base:   &base.Checker{},
	}
}

// validateTx checks whether a transaction is valid according to the consensus
// rules
func (v *Checker) validateTx(tx *types.Transaction) error {
	// Transactions can't be negative. This may never happen using RLP decoded
	// transactions but may occur if you create a transaction using the RPC.
	if tx.Value().Sign() < 0 || tx.GasPrice().Sign() < 0 {
		return ErrNegativeValue
	}
	// Ensure the transaction has more gas than the basic tx fee.
	intrGas, err := evmcore.IntrinsicGas(tx.Data(), tx.To() == nil, true)
	if err != nil {
		return err
	}
	if tx.Gas() < intrGas {
		return ErrIntrinsicGas
	}
	if tx.GasPrice().Cmp(params.MinGasPrice) < 0 {
		return ErrUnderpriced
	}
	return nil
}

func (v *Checker) checkTxs(e inter.EventPayloadI) error {
	for _, tx := range e.Txs() {
		if err := v.validateTx(tx); err != nil {
			return err
		}
	}
	return nil
}

func CalcGasPowerUsed(e inter.EventPayloadI, config *opera.DagConfig) uint64 {
	txsGas := uint64(0)
	for _, tx := range e.Txs() {
		txsGas += tx.Gas()
	}

	parentsGas := uint64(0)
	if len(e.Parents()) > config.MaxFreeParents {
		parentsGas = uint64(len(e.Parents())-config.MaxFreeParents) * params.ParentGas
	}
	extraGas := uint64(len(e.Extra())) * params.ExtraDataGas

	return txsGas + parentsGas + extraGas + params.EventGas
}

func (v *Checker) checkGas(e inter.EventPayloadI) error {
	if e.GasPowerUsed() > params.MaxGasPowerUsed {
		return ErrTooBigGasUsed
	}
	if e.GasPowerUsed() != CalcGasPowerUsed(e, v.config) {
		return ErrWrongGasUsed
	}

	return nil
}

// Validate event
func (v *Checker) Validate(e inter.EventPayloadI) error {
	if err := v.base.Validate(e); err != nil {
		return err
	}
	if e.GasPowerUsed() >= math.MaxInt64-1 || e.GasPowerLeft().Max() >= math.MaxInt64-1 {
		return base.ErrHugeValue
	}
	if e.CreationTime() <= 0 || e.MedianTime() <= 0 {
		return ErrZeroTime
	}
	if len(e.Parents()) > v.config.MaxParents {
		return ErrTooManyParents
	}
	if err := v.checkGas(e); err != nil {
		return err
	}
	if err := v.checkTxs(e); err != nil {
		return err
	}

	return nil
}
