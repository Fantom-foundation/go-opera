package eventcheck

import (
	"github.com/Fantom-foundation/go-lachesis/eventcheck/basiccheck"
	"github.com/Fantom-foundation/go-lachesis/eventcheck/epochcheck"
	"github.com/Fantom-foundation/go-lachesis/eventcheck/gaspowercheck"
	"github.com/Fantom-foundation/go-lachesis/eventcheck/heavycheck"
	"github.com/Fantom-foundation/go-lachesis/eventcheck/parentscheck"
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/lachesis"
	"github.com/ethereum/go-ethereum/core/types"
)

type DagReader interface {
	epochcheck.DagReader
	gaspowercheck.DagReader
}

// ValidateAll runs all the checks except Poset-related. intended only for tests
func ValidateAll(config *lachesis.DagConfig, reader epochcheck.DagReader, txSigner types.Signer, e *inter.Event, parents []*inter.EventHeaderData) error {
	if err := basiccheck.New(config).Validate(e); err != nil {
		return err
	}
	if err := epochcheck.New(config, reader).Validate(e); err != nil {
		return err
	}
	if err := parentscheck.New(config).Validate(e, parents); err != nil {
		return err
	}
	if err := heavycheck.New(config, txSigner, 1).Validate(e); err != nil {
		return err
	}
	return nil
}
