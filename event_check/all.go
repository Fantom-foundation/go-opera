package event_check

import (
	"github.com/Fantom-foundation/go-lachesis/event_check/basic_check"
	"github.com/Fantom-foundation/go-lachesis/event_check/epoch_check"
	"github.com/Fantom-foundation/go-lachesis/event_check/heavy_check"
	"github.com/Fantom-foundation/go-lachesis/event_check/parents_check"
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/lachesis"
	"github.com/ethereum/go-ethereum/core/types"
)

// ValidateAll runs all the checks except Poset-related. intended only for tests
func ValidateAll(config *lachesis.DagConfig, reader epoch_check.DagReader, txSigner types.Signer, e *inter.Event, parents []*inter.EventHeaderData) error {
	if err := basic_check.New(config).Validate(e); err != nil {
		return err
	}
	if err := epoch_check.New(config, reader).Validate(e); err != nil {
		return err
	}
	if err := parents_check.New(config).Validate(e, parents); err != nil {
		return err
	}
	if err := heavy_check.New(config, txSigner, 1).Validate(e); err != nil {
		return err
	}
	return nil
}
