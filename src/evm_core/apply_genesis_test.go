package evm_core

import (
	"testing"

	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/src/kvdb/memorydb"
	"github.com/Fantom-foundation/go-lachesis/src/kvdb/no_key_is_err"
	"github.com/Fantom-foundation/go-lachesis/src/kvdb/table"
	"github.com/Fantom-foundation/go-lachesis/src/lachesis/genesis"
	"github.com/Fantom-foundation/go-lachesis/src/logger"
)

func TestApplyGenesis(t *testing.T) {
	assertar := assert.New(t)
	logger.SetTestMode(t)

	db := rawdb.NewDatabase(
		no_key_is_err.Wrap(
			table.New(
				memorydb.New(), []byte("evm_"))))

	// no genesis
	_, err := ApplyGenesis(db, nil)
	if !assertar.Error(err) {
		return
	}

	// the same genesis
	genA := genesis.FakeGenesis(3)
	blockA1, err := ApplyGenesis(db, &genA)
	if !assertar.NoError(err) {
		return
	}
	blockA2, err := ApplyGenesis(db, &genA)
	if !assertar.NoError(err) {
		return
	}
	if !assertar.Equal(blockA1, blockA2) {
		return
	}

	// different genesis
	genB := genesis.FakeGenesis(3)
	_, err = ApplyGenesis(db, &genB)
	if !assertar.Error(err) {
		return
	}

}
