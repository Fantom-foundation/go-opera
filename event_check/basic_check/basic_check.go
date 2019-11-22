package basic_check

import (
	"errors"
	"math"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"

	"github.com/Fantom-foundation/go-lachesis/evm_core"
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/lachesis"
)

const (
	MaxGasPowerUsed = params.GenesisGasLimit * 3
	// MaxEventSize ensures that in all the "real" cases, the event will be limited by gas, not size.
	// Yet it's technically possible to construct an event which is limited by size.
	MaxEventSize = MaxGasPowerUsed / params.TxDataNonZeroGas
	MaxExtraData = 256 // it has fair gas cost, so it's fine to have a high limit

	EventGas  = params.TxGas // TODO estimate the cost more accurately
	ParentGas = EventGas / 5
	// ExtraDataGas is cost per byte of extra event data. It's higher than regular data price, because it's a part of the header
	ExtraDataGas = params.TxDataNonZeroGas * 2
)

var (
	ErrSigMalformed   = errors.New("event signature malformed")
	ErrVersion        = errors.New("event has wrong version")
	ErrTooLarge       = errors.New("event size exceeds the limit")
	ErrExtraTooLarge  = errors.New("event extra is too big")
	ErrNoParents      = errors.New("event has no parents")
	ErrTooManyParents = errors.New("event has too many parents")
	ErrTooBigGasUsed  = errors.New("event uses too much gas power")
	ErrWrongGasUsed   = errors.New("event has incorrect gas power")
	ErrIntrinsicGas   = errors.New("intrinsic gas too low")
	ErrNotInited      = errors.New("event field is not initialized")
	ErrZeroTime       = errors.New("event has zero timestamp")
	ErrNegativeValue  = errors.New("negative value")
	ErrHugeValue      = errors.New("too big value")
)

type Validator struct {
	config *lachesis.DagConfig
}

// New validator which performs checks which don't require anything except event
func New(config *lachesis.DagConfig) *Validator {
	return &Validator{
		config: config,
	}
}

// validateTx checks whether a transaction is valid according to the consensus
// rules
func (v *Validator) validateTx(tx *types.Transaction) error {
	// Transactions can't be negative. This may never happen using RLP decoded
	// transactions but may occur if you create a transaction using the RPC.
	if tx.Value().Sign() < 0 || tx.GasPrice().Sign() < 0 {
		return ErrNegativeValue
	}
	// Ensure the transaction has more gas than the basic tx fee.
	intrGas, err := evm_core.IntrinsicGas(tx.Data(), tx.To() == nil, true)
	if err != nil {
		return err
	}
	if tx.Gas() < intrGas {
		return ErrIntrinsicGas
	}
	return nil
}

func (v *Validator) checkTxs(e *inter.Event) error {
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
		parentsGas = uint64(len(e.Parents)-config.MaxFreeParents) * ParentGas
	}
	extraGas := uint64(len(e.Extra)) * ExtraDataGas

	return txsGas + parentsGas + extraGas + EventGas
}

func (v *Validator) checkGas(e *inter.Event) error {
	if e.GasPowerUsed > MaxGasPowerUsed {
		return ErrTooBigGasUsed
	}
	if e.GasPowerUsed != CalcGasPowerUsed(e, v.config) {
		return ErrWrongGasUsed
	}

	return nil
}

func (v *Validator) checkLimits(e *inter.Event) error {
	if uint64(e.Size()) > MaxEventSize {
		return ErrTooLarge
	}
	if len(e.Extra) > MaxExtraData {
		return ErrExtraTooLarge
	}
	if len(e.Parents) > v.config.MaxParents {
		return ErrTooManyParents
	}
	if e.Seq >= math.MaxInt32/2 || e.Epoch >= math.MaxInt32/2 || e.Frame >= math.MaxInt32/2 ||
		e.Lamport >= math.MaxInt32/2 || e.GasPowerUsed >= math.MaxInt64/2 || e.GasPowerLeft >= math.MaxInt64/2 {
		return ErrHugeValue
	}

	return nil
}

func (v *Validator) checkInited(e *inter.Event) error {
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

func (v *Validator) Validate(e *inter.Event) error {
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
