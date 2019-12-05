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
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// NewTxsNotify is posted when a batch of transactions enter the transaction pool.
type NewTxsNotify struct{ Txs []*types.Transaction }

// PendingLogsNotify is posted pre mining and notifies of pending logs.
type PendingLogsNotify struct {
	Logs []*types.Log
}

// NewMinedBlockNotify is posted when a block has been imported.
type NewMinedBlockNotify struct{ Block *EvmBlock }

// RemovedLogsNotify is posted when a reorg happens
type RemovedLogsNotify struct{ Logs []*types.Log }

type ChainNotify struct {
	Block *EvmBlock
	Hash  common.Hash
	Logs  []*types.Log
}

type ChainSideNotify struct {
	Block *EvmBlock
}

type ChainHeadNotify struct{ Block *EvmBlock }
