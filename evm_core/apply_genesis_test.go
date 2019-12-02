package evm_core

import (
	"testing"

	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/inter/pos"
	"github.com/Fantom-foundation/go-lachesis/kvdb/memorydb"
	"github.com/Fantom-foundation/go-lachesis/kvdb/no_key_is_err"
	"github.com/Fantom-foundation/go-lachesis/kvdb/table"
	"github.com/Fantom-foundation/go-lachesis/lachesis"
	"github.com/Fantom-foundation/go-lachesis/lachesis/genesis"
	"github.com/Fantom-foundation/go-lachesis/logger"
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
	accsA := genesis.FakeAccounts(0, 3, 1e6*pos.Qualification)
	netA := lachesis.FakeNetConfig(accsA)
	blockA1, err := ApplyGenesis(db, &netA)
	if !assertar.NoError(err) {
		return
	}
	blockA2, err := ApplyGenesis(db, &netA)
	if !assertar.NoError(err) {
		return
	}
	if !assertar.Equal(blockA1, blockA2) {
		return
	}

	// different genesis
	accsB := genesis.FakeAccounts(0, 4, 1e6*pos.Qualification)
	netB := lachesis.FakeNetConfig(accsB)
	_, err = ApplyGenesis(db, &netB)
	if !assertar.Error(err) {
		return
	}

}
