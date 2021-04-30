package epochcheck

import (
	"errors"
	"strings"

	base "github.com/Fantom-foundation/lachesis-base/eventcheck/epochcheck"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"

	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/opera"
)

var (
	ErrTooManyParents = errors.New("event has too many parents")
	ErrTooBigGasUsed  = errors.New("event uses too much gas power")
	ErrWrongGasUsed   = errors.New("event has incorrect gas power")
	ErrUnderpriced    = errors.New("event transaction underpriced")
	ErrTooBigExtra    = errors.New("event extra data is too large")
	ErrNotRelevant    = base.ErrNotRelevant
	ErrAuth           = base.ErrAuth
)

// Reader returns currents epoch and its validators group.
type Reader interface {
	base.Reader
	GetEpochRules() (opera.Rules, idx.Epoch)
}

// Checker which require only current epoch info
type Checker struct {
	Base   *base.Checker
	reader Reader
}

func New(reader Reader) *Checker {
	return &Checker{
		Base:   base.New(reader),
		reader: reader,
	}
}

func CalcGasPowerUsed(e inter.EventPayloadI, rules opera.Rules) uint64 {
	txsGas := uint64(0)
	for _, tx := range e.Txs() {
		txsGas += tx.Gas()
	}

	parentsGas := uint64(0)
	if idx.Event(len(e.Parents())) > rules.Dag.MaxFreeParents {
		parentsGas = uint64(idx.Event(len(e.Parents()))-rules.Dag.MaxFreeParents) * rules.Economy.Gas.ParentGas
	}
	extraGas := uint64(len(e.Extra())) * rules.Economy.Gas.ExtraDataGas

	return txsGas + parentsGas + extraGas + rules.Economy.Gas.EventGas
}

func (v *Checker) checkGas(e inter.EventPayloadI, rules opera.Rules) error {
	if e.GasPowerUsed() > rules.Economy.Gas.MaxEventGas {
		return ErrTooBigGasUsed
	}
	if e.GasPowerUsed() != CalcGasPowerUsed(e, rules) {
		return ErrWrongGasUsed
	}
	return nil
}

// Validate event
func (v *Checker) Validate(e inter.EventPayloadI) error {
	if err := v.Base.Validate(e); err != nil {
		if strings.HasSuffix(err.Error(), "no space left on device") {
			panic("HERE x")
		}
		return err
	}
	rules, epoch := v.reader.GetEpochRules()
	// Check epoch of the rules to prevent a race condition
	if e.Epoch() != epoch {
		return base.ErrNotRelevant
	}
	if idx.Event(len(e.Parents())) > rules.Dag.MaxParents {
		return ErrTooManyParents
	}
	if uint32(len(e.Extra())) > rules.Dag.MaxExtraData {
		return ErrTooBigExtra
	}
	if err := v.checkGas(e, rules); err != nil {
		if strings.HasSuffix(err.Error(), "no space left on device") {
			panic("HERE x")
		}
		return err
	}
	for _, tx := range e.Txs() {
		if tx.GasPrice().Cmp(rules.Economy.MinGasPrice) < 0 {
			return ErrUnderpriced
		}
	}
	return nil
}
