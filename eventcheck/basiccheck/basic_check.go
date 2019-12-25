package basiccheck

import (
	"errors"
	"math"

	"github.com/ethereum/go-ethereum/core/types"

	"github.com/Fantom-foundation/go-lachesis/evmcore"
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/lachesis"
	"github.com/Fantom-foundation/go-lachesis/lachesis/params"
)

var (
	ErrSigMalformed   = errors.New("event signature malformed")
	ErrVersion        = errors.New("event has wrong version")
	ErrExtraTooLarge  = errors.New("event extra is too big")
	ErrNoParents      = errors.New("event has no parents")
	ErrTooManyParents = errors.New("event has too many parents")
	ErrTooBigGasUsed  = errors.New("event uses too much gas power")
	ErrWrongGasUsed   = errors.New("event has incorrect gas power")
	ErrIntrinsicGas   = errors.New("intrinsic gas too low")
	ErrUnderpriced    = errors.New("event transaction underpriced")
	ErrNotInited      = errors.New("event field is not initialized")
	ErrZeroTime       = errors.New("event has zero timestamp")
	ErrNegativeValue  = errors.New("negative value")
	ErrHugeValue      = errors.New("too big value")
)

type Checker struct {
	config *lachesis.DagConfig
}

// New validator which performs checks which don't require anything except event
func New(config *lachesis.DagConfig) *Checker {
	return &Checker{
		config: config,
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

func (v *Checker) checkTxs(e *inter.Event) error {
	for _, tx := range e.Transactions {
		if err := v.validateTx(tx); err != nil {
			return err
		}
	}
	return nil
}

func CalcGasPowerUsed(e *inter.Event, config *lachesis.DagConfig) uint64 {
	txsGas := uint64(0)
	for _, tx := range e.Transactions {
		txsGas += tx.Gas()
	}

	parentsGas := uint64(0)
	if len(e.Parents) > config.MaxFreeParents {
		parentsGas = uint64(len(e.Parents)-config.MaxFreeParents) * params.ParentGas
	}
	extraGas := uint64(len(e.Extra)) * params.ExtraDataGas

	return txsGas + parentsGas + extraGas + params.EventGas
}

func (v *Checker) checkGas(e *inter.Event) error {
	if e.GasPowerUsed > params.MaxGasPowerUsed {
		return ErrTooBigGasUsed
	}
	if e.GasPowerUsed != CalcGasPowerUsed(e, v.config) {
		return ErrWrongGasUsed
	}

	return nil
}

func (v *Checker) checkLimits(e *inter.Event) error {
	if len(e.Extra) > params.MaxExtraData {
		return ErrExtraTooLarge
	}
	if len(e.Parents) > v.config.MaxParents {
		return ErrTooManyParents
	}
	if e.Seq >= math.MaxInt32/2 || e.Epoch >= math.MaxInt32/2 || e.Frame >= math.MaxInt32/2 ||
		e.Lamport >= math.MaxInt32/2 || e.GasPowerUsed >= math.MaxInt64/2 || e.GasPowerLeft.Max() >= math.MaxInt64/2 {
		return ErrHugeValue
	}

	return nil
}

func (v *Checker) checkInited(e *inter.Event) error {
	if e.Seq <= 0 || e.Epoch <= 0 || e.Frame <= 0 || e.Lamport <= 0 {
		return ErrNotInited // it's unsigned, but check for negative in a case if type will change
	}

	if e.ClaimedTime <= 0 {
		return ErrZeroTime
	}
	if e.Seq > 1 && len(e.Parents) == 0 {
		return ErrNoParents
	}
	if len(e.Sig) != 65 {
		return ErrSigMalformed
	}

	return nil
}

// Validate event
func (v *Checker) Validate(e *inter.Event) error {
	if e.Version != 0 {
		return ErrVersion
	}
	if err := v.checkLimits(e); err != nil {
		return err
	}
	if err := v.checkInited(e); err != nil {
		return err
	}
	if err := v.checkGas(e); err != nil {
		return err
	}
	if err := v.checkTxs(e); err != nil {
		return err
	}

	return nil
}
