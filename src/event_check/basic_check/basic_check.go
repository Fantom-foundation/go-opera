package basic_check

import (
	"errors"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"

	"github.com/Fantom-foundation/go-lachesis/src/evm_core"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/lachesis"
)

const (
	MaxGasPowerUsed = params.GenesisGasLimit
	// It ensures that in all the "real" cases, the event will be limited by gas, not size.
	// Yet it's technically possible to construct an event which is limited by size.
	MaxEventSize = MaxGasPowerUsed / params.TxDataNonZeroGas
	MaxExtraData = 256 // it has fair gas price, so it's fine to have a high limit

	EventGas  = params.TxGas // TODO estimate the cost more accurately
	ParentGas = EventGas / 5
	// Per byte of extra event data. It's higher than regular data price, because it's a part of the header
	ExtraDataGas = params.TxDataNonZeroGas * 2
)

var (
	ErrSigMalformed   = errors.New("event signature malformed")
	ErrVersion        = errors.New("event has wrong version")
	ErrTooLarge       = errors.New("event size exceeds the limit")
	ErrExtraTooLarge  = errors.New("event extra is too big")
	ErrNoParents      = errors.New("event has no parents")
	ErrTooMuchParents = errors.New("event has too much parents")
	ErrTooBigGasUsed  = errors.New("event uses too much gas power")
	ErrWrongGasUsed   = errors.New("event has incorrect gas power")
	ErrIntrinsicGas   = errors.New("intrinsic gas too low")
	ErrMalformed      = errors.New("event is malformed")
	ErrZeroTime       = errors.New("event has zero timestamp")
	ErrNegativeValue  = errors.New("negative value")
)

// Check which don't require anything except event
type Validator struct {
	config *lachesis.DagConfig
}

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

func CalcGasPowerUsed(e *inter.Event) uint64 {
	txsGas := uint64(0)
	for _, tx := range e.Transactions {
		txsGas += tx.Gas()
	}

	parentsGas := uint64(len(e.Parents)) * ParentGas
	extraGas := uint64(len(e.Extra)) * ExtraDataGas

	return txsGas + parentsGas + extraGas + EventGas
}

func (v *Validator) checkGas(e *inter.Event) error {
	if e.GasPowerUsed > MaxGasPowerUsed {
		return ErrTooBigGasUsed
	}
	if e.GasPowerUsed != CalcGasPowerUsed(e) {
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
		return ErrTooMuchParents
	}
	return nil
}

func (v *Validator) checkInited(e *inter.Event) error {
	if e.Seq == 0 || e.Epoch == 0 || e.Frame == 0 || e.Lamport == 0 {
		return ErrMalformed
	}
	if e.ClaimedTime == 0 {
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
