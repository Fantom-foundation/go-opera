// Copyright 2014 The go-ethereum Authors
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

package filters

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	notify "github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/Fantom-foundation/go-opera/evmcore"
	"github.com/Fantom-foundation/go-opera/gossip/evmstore"
	"github.com/Fantom-foundation/go-opera/topicsdb"
)

type Backend interface {
	HeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*evmcore.EvmHeader, error)
	HeaderByHash(ctx context.Context, blockHash common.Hash) (*evmcore.EvmHeader, error)
	GetReceipts(ctx context.Context, blockHash common.Hash) (types.Receipts, error)
	GetReceiptsByNumber(ctx context.Context, number rpc.BlockNumber) (types.Receipts, error)
	GetLogs(ctx context.Context, blockHash common.Hash) ([][]*types.Log, error)
	GetTxPosition(txid common.Hash) *evmstore.TxPosition

	SubscribeNewBlockNotify(ch chan<- evmcore.ChainHeadNotify) notify.Subscription
	SubscribeNewTxsNotify(chan<- evmcore.NewTxsNotify) notify.Subscription
	SubscribeLogsNotify(ch chan<- []*types.Log) notify.Subscription

	EvmLogIndex() topicsdb.Index

	CalcBlockExtApi() bool
}

// Filter can be used to retrieve and filter logs.
type Filter struct {
	backend Backend
	config  Config

	addresses []common.Address
	topics    [][]common.Hash

	block      common.Hash // Block hash if filtering a single block
	begin, end int64       // Range interval if filtering multiple blocks
}

// NewRangeFilter creates a new filter which inspects the blocks to
// figure out whether a particular block is interesting or not.
func NewRangeFilter(backend Backend, cfg Config, begin, end int64, addresses []common.Address, topics [][]common.Hash) *Filter {
	// Create a generic filter and convert it into a range filter
	filter := newFilter(backend, cfg, addresses, topics)

	filter.begin = begin
	filter.end = end

	return filter
}

// NewBlockFilter creates a new filter which directly inspects the contents of
// a block to figure out whether it is interesting or not.
func NewBlockFilter(backend Backend, cfg Config, block common.Hash, addresses []common.Address, topics [][]common.Hash) *Filter {
	// Create a generic filter and convert it into a block filter
	filter := newFilter(backend, cfg, addresses, topics)

	filter.block = block

	return filter
}

// newFilter creates a generic filter that can either filter based on a block hash,
// or based on range queries. The search criteria needs to be explicitly set.
func newFilter(backend Backend, cfg Config, addresses []common.Address, topics [][]common.Hash) *Filter {
	return &Filter{
		backend:   backend,
		config:    cfg,
		addresses: addresses,
		topics:    topics,
	}
}

// Logs searches the blockchain for matching log entries, returning all from the
// first block that contains matches, updating the start of the filter accordingly.
func (f *Filter) Logs(ctx context.Context) ([]*types.Log, error) {
	// If we're doing singleton block filtering, execute and return
	if f.block != common.Hash(hash.Zero) {
		header, err := f.backend.HeaderByHash(ctx, f.block)
		if err != nil {
			return nil, err
		}
		if header == nil {
			return nil, errors.New("unknown block")
		}
		return f.blockLogs(ctx, header.Hash)
	}
	// Figure out the limits of the filter range
	header, _ := f.backend.HeaderByNumber(ctx, rpc.LatestBlockNumber)
	if header == nil {
		return nil, nil
	}
	head := idx.Block(header.Number.Uint64())

	begin := idx.Block(f.begin)
	if f.begin < 0 {
		begin = head
	}
	end := idx.Block(f.end)
	if f.end < 0 {
		end = head
	}
	if begin > end {
		return []*types.Log{}, nil
	}

	if isEmpty(f.topics) && len(f.addresses) == 0 {
		return f.unindexedLogs(ctx, begin, end)
	} else {
		return f.indexedLogs(ctx, begin, end)
	}
}

// indexedLogs returns the logs matching the filter criteria based on topics index.
func (f *Filter) indexedLogs(ctx context.Context, begin, end idx.Block) ([]*types.Log, error) {
	if end-begin > f.config.IndexedLogsBlockRangeLimit {
		return nil, fmt.Errorf("too wide blocks range, the limit is %d", f.config.IndexedLogsBlockRangeLimit)
	}

	addresses := make([]common.Hash, len(f.addresses))
	for i, addr := range f.addresses {
		addresses[i] = addr.Hash()
	}

	pattern := make([][]common.Hash, 1, len(f.topics)+1)
	pattern[0] = addresses
	pattern = append(pattern, f.topics...)

	logs, err := f.backend.EvmLogIndex().FindInBlocks(ctx, begin, end, pattern)
	if err != nil {
		return nil, err
	}

	for _, l := range logs {
		pos := f.backend.GetTxPosition(l.TxHash)
		if pos != nil {
			l.TxIndex = uint(pos.BlockOffset)
		} else {
			log.Warn("tx index empty", "hash", l.TxHash)
		}
	}

	return logs, nil
}

// indexedLogs returns the logs matching the filter criteria based on raw block
// iteration.
func (f *Filter) unindexedLogs(ctx context.Context, begin, end idx.Block) (logs []*types.Log, err error) {
	if end-begin > f.config.UnindexedLogsBlockRangeLimit {
		return nil, fmt.Errorf("too wide blocks range, the limit is %d", f.config.UnindexedLogsBlockRangeLimit)
	}

	var (
		header *evmcore.EvmHeader
		found  []*types.Log
	)
	for n := begin; n <= end; n++ {
		err = ctx.Err()
		if err != nil {
			return
		}

		header, err = f.backend.HeaderByNumber(ctx, rpc.BlockNumber(n))
		if header == nil || err != nil {
			return
		}
		found, err = f.blockLogs(ctx, header.Hash)
		if err != nil {
			return
		}
		logs = append(logs, found...)
	}
	return
}

// blockLogs returns the logs matching the filter criteria within a single block.
func (f *Filter) blockLogs(ctx context.Context, header common.Hash) ([]*types.Log, error) {
	// Get the logs of the block
	logsList, err := f.backend.GetLogs(ctx, header)
	if err != nil {
		return nil, err
	}

	var unfiltered []*types.Log
	for _, logs := range logsList {
		unfiltered = append(unfiltered, logs...)
	}

	logs := filterLogs(unfiltered, nil, nil, f.addresses, f.topics)
	if len(logs) > 0 {
		// We have matching logs, check if we need to resolve full logs via the light client
		if logs[0].TxHash == common.Hash(hash.Zero) {
			receipts, err := f.backend.GetReceipts(ctx, header)
			if err != nil {
				return nil, err
			}
			unfiltered = unfiltered[:0]
			for _, receipt := range receipts {
				unfiltered = append(unfiltered, receipt.Logs...)
			}
			logs = filterLogs(unfiltered, nil, nil, f.addresses, f.topics)
		}
		return logs, nil
	}
	return nil, nil
}

func includes(addresses []common.Address, a common.Address) bool {
	for _, addr := range addresses {
		if addr == a {
			return true
		}
	}

	return false
}

// filterLogs creates a slice of logs matching the given criteria.
func filterLogs(logs []*types.Log, fromBlock, toBlock *big.Int, addresses []common.Address, topics [][]common.Hash) []*types.Log {
	var ret []*types.Log
Logs:
	for _, log := range logs {
		if fromBlock != nil && fromBlock.Int64() >= 0 && fromBlock.Uint64() > log.BlockNumber {
			continue
		}
		if toBlock != nil && toBlock.Int64() >= 0 && toBlock.Uint64() < log.BlockNumber {
			continue
		}

		if len(addresses) > 0 && !includes(addresses, log.Address) {
			continue
		}
		// If the to filtered topics is greater than the amount of topics in logs, skip.
		if len(topics) > len(log.Topics) {
			continue Logs
		}
		for i, sub := range topics {
			match := len(sub) == 0 // empty rule set == wildcard
			for _, topic := range sub {
				if log.Topics[i] == topic {
					match = true
					break
				}
			}
			if !match {
				continue Logs
			}
		}
		ret = append(ret, log)
	}
	return ret
}

func isEmpty(topics [][]common.Hash) bool {
	for _, tt := range topics {
		if len(tt) > 0 {
			return false
		}
	}
	return true
}
