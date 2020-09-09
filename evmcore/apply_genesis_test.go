// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package evmcore

import (
	"math/big"
	"testing"

	"github.com/Fantom-foundation/lachesis-base/kvdb/memorydb"
	"github.com/Fantom-foundation/lachesis-base/kvdb/nokeyiserr"
	"github.com/Fantom-foundation/lachesis-base/kvdb/table"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-opera/logger"
	"github.com/Fantom-foundation/go-opera/opera"
	"github.com/Fantom-foundation/go-opera/opera/genesis"
	"github.com/Fantom-foundation/go-opera/utils/adapters/kvdb2ethdb"
)

func TestApplyGenesis(t *testing.T) {
	assertar := assert.New(t)
	logger.SetTestMode(t)

	db1 := rawdb.NewDatabase(
		kvdb2ethdb.Wrap(
			nokeyiserr.Wrap(
				table.New(
					memorydb.New(), []byte("evm1_")))))
	db2 := rawdb.NewDatabase(
		kvdb2ethdb.Wrap(
			nokeyiserr.Wrap(
				table.New(
					memorydb.New(), []byte("evm2_")))))

	// no genesis
	_, err := ApplyGenesis(db1, nil)
	if !assertar.Error(err) {
		return
	}

	// the same genesis
	accsA := genesis.FakeValidators(3, big.NewInt(10000000000), big.NewInt(1))
	netA := opera.FakeNetConfig(accsA)
	blockA1, err := ApplyGenesis(db1, &netA)
	if !assertar.NoError(err) {
		return
	}
	blockA2, err := ApplyGenesis(db2, &netA)
	if !assertar.NoError(err) {
		return
	}
	if !assertar.Equal(blockA1, blockA2) {
		return
	}

	// different genesis
	accsB := genesis.FakeValidators(4, big.NewInt(10000000000), big.NewInt(1))
	netB := opera.FakeNetConfig(accsB)
	blockB, err := ApplyGenesis(db2, &netB)
	if !assertar.NotEqual(blockA1, blockB) {
		return
	}
	if !assertar.NoError(err) {
		return
	}

}
