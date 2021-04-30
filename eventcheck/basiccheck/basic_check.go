package basiccheck

import (
	"errors"
	"math"
	"strings"

	base "github.com/Fantom-foundation/lachesis-base/eventcheck/basiccheck"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/Fantom-foundation/go-opera/evmcore"
	"github.com/Fantom-foundation/go-opera/inter"
)

var (
	ErrZeroTime      = errors.New("event has zero timestamp")
	ErrNegativeValue = errors.New("negative value")
	ErrIntrinsicGas  = errors.New("intrinsic gas too low")
)

type Checker struct {
	base base.Checker
}

// New validator which performs checks which don't require anything except event
func New() *Checker {
	return &Checker{
		base: base.Checker{},
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
	intrGas, err := evmcore.IntrinsicGas(tx.Data(), tx.To() == nil)
	if err != nil {
		return err
	}
	if tx.Gas() < intrGas {
		return ErrIntrinsicGas
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

// Validate event
func (v *Checker) Validate(e inter.EventPayloadI) error {
	if err := v.base.Validate(e); err != nil {
		if strings.HasSuffix(err.Error(), "no space left on device") {
			panic("HERE x")
		}
		return err
	}
	if e.GasPowerUsed() >= math.MaxInt64-1 || e.GasPowerLeft().Max() >= math.MaxInt64-1 {
		return base.ErrHugeValue
	}
	if e.CreationTime() <= 0 || e.MedianTime() <= 0 {
		return ErrZeroTime
	}
	if err := v.checkTxs(e); err != nil {
		if strings.HasSuffix(err.Error(), "no space left on device") {
			panic("HERE x")
		}
		return err
	}

	return nil
}
