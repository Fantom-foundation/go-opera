// Copyright 2016 The go-ethereum Authors
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
	"math/big"
	"reflect"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethdb"
	notify "github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/Fantom-foundation/go-lachesis/evmcore"
	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter/pos"
	"github.com/Fantom-foundation/go-lachesis/kvdb/memorydb"
	"github.com/Fantom-foundation/go-lachesis/lachesis"
	"github.com/Fantom-foundation/go-lachesis/lachesis/genesis"
	"github.com/Fantom-foundation/go-lachesis/topicsdb"
)

type testBackend struct {
	db         ethdb.Database
	logIndex   *topicsdb.Index
	blocksFeed *notify.Feed
	txsFeed    *notify.Feed
	logsFeed   *notify.Feed
}

func newTestBackend() *testBackend {
	return &testBackend{
		db:         rawdb.NewMemoryDatabase(),
		logIndex:   topicsdb.New(memorydb.New()),
		blocksFeed: new(notify.Feed),
		txsFeed:    new(notify.Feed),
		logsFeed:   new(notify.Feed),
	}
}

func (b *testBackend) ChainDb() ethdb.Database {
	return b.db
}

func (b *testBackend) HeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*evmcore.EvmHeader, error) {
	var (
		hash common.Hash
		num  uint64
	)
	if blockNr == rpc.LatestBlockNumber {
		hash = rawdb.ReadHeadBlockHash(b.db)
		number := rawdb.ReadHeaderNumber(b.db, hash)
		if number == nil {
			return nil, nil
		}
		num = *number
	} else {
		num = uint64(blockNr)
		hash = rawdb.ReadCanonicalHash(b.db, num)
	}

	h := rawdb.ReadHeader(b.db, hash, num)
	eh := evmcore.ConvertFromEthHeader(h)
	eh.Hash = h.Hash()
	return eh, nil
}

func (b *testBackend) HeaderByHash(ctx context.Context, hash common.Hash) (*evmcore.EvmHeader, error) {
	number := rawdb.ReadHeaderNumber(b.db, hash)
	if number == nil {
		return nil, nil
	}

	h := rawdb.ReadHeader(b.db, hash, *number)
	eh := evmcore.ConvertFromEthHeader(h)
	eh.Hash = h.Hash()
	return eh, nil
}

func (b *testBackend) GetReceipts(ctx context.Context, hash common.Hash) (types.Receipts, error) {
	if number := rawdb.ReadHeaderNumber(b.db, hash); number != nil {
		return rawdb.ReadReceipts(b.db, hash, *number, params.TestChainConfig), nil
	}
	return nil, nil
}

func (b *testBackend) GetLogs(ctx context.Context, hash common.Hash) ([][]*types.Log, error) {
	number := rawdb.ReadHeaderNumber(b.db, hash)
	if number == nil {
		return nil, nil
	}

	receipts := rawdb.ReadReceipts(b.db, hash, *number, params.TestChainConfig)
	logs := make([][]*types.Log, len(receipts))
	for i, receipt := range receipts {
		logs[i] = receipt.Logs
	}

	return logs, nil
}

func (b *testBackend) SubscribeNewTxsEvent(ch chan<- core.NewTxsEvent) notify.Subscription {
	return b.txsFeed.Subscribe(ch)
}

func (b *testBackend) SubscribeLogsEvent(ch chan<- []*types.Log) notify.Subscription {
	return b.logsFeed.Subscribe(ch)
}

func (b *testBackend) SubscribeNewBlockEvent(ch chan<- evmcore.ChainHeadNotify) notify.Subscription {
	return b.blocksFeed.Subscribe(ch)
}

func (b *testBackend) EvmLogIndex() *topicsdb.Index {
	return b.logIndex
}

// TestBlockSubscription tests if a block subscription returns block hashes for posted chain notify.
// It creates multiple subscriptions:
// - one at the start and should receive all posted chain events and a second (blockHashes)
// - one that is created after a cutoff moment and uninstalled after a second cutoff moment (blockHashes[cutoff1:cutoff2])
// - one that is created after the second cutoff moment (blockHashes[cutoff2:])
func TestBlockSubscription(t *testing.T) {
	t.Parallel()

	var (
		backend = newTestBackend()
		api     = NewPublicFilterAPI(backend)

		net         = lachesis.FakeNetConfig(genesis.FakeValidators(5, big.NewInt(0), pos.StakeToBalance(1)))
		genesis     = evmcore.MustApplyGenesis(&net, backend.db)
		chain, _, _ = evmcore.GenerateChain(
			params.TestChainConfig, genesis, backend.db, 10, nil)
		chainEvents = []evmcore.ChainHeadNotify{}
	)

	for _, blk := range chain {
		chainEvents = append(chainEvents, evmcore.ChainHeadNotify{
			Block: &evmcore.EvmBlock{
				EvmHeader:    blk.EvmHeader,
				Transactions: blk.Transactions,
			},
		})
	}

	chan0 := make(chan *types.Header)
	sub0 := api.events.SubscribeNewHeads(chan0)
	chan1 := make(chan *types.Header)
	sub1 := api.events.SubscribeNewHeads(chan1)

	go func() { // simulate client
		i1, i2 := 0, 0
		for i1 != len(chainEvents) || i2 != len(chainEvents) {
			select {
			case header := <-chan0:
				got := hash.FromBytes(header.Extra)
				if chainEvents[i1].Block.Hash != got {
					t.Errorf("sub0 received invalid hash on index %d, want %x, got %x", i1, chainEvents[i1].Block.Hash, got)
				}
				i1++
			case header := <-chan1:
				got := hash.FromBytes(header.Extra)
				if chainEvents[i2].Block.Hash != got {
					t.Errorf("sub1 received invalid hash on index %d, want %x, got %x", i2, chainEvents[i2].Block.Hash, got)
				}
				i2++
			}
		}

		sub0.Unsubscribe()
		sub1.Unsubscribe()
	}()

	time.Sleep(1 * time.Second)
	for _, e := range chainEvents {
		backend.blocksFeed.Send(e)
	}

	<-sub0.Err()
	<-sub1.Err()
}

// TestPendingTxFilter tests whether pending tx filters retrieve all pending transactions that are posted to the event mux.
func TestPendingTxFilter(t *testing.T) {
	t.Parallel()

	var (
		backend = newTestBackend()
		api     = NewPublicFilterAPI(backend)

		transactions = []*types.Transaction{
			types.NewTransaction(0, common.HexToAddress("0xb794f5ea0ba39494ce83a213fffba74279579268"), new(big.Int), 0, new(big.Int), nil),
			types.NewTransaction(1, common.HexToAddress("0xb794f5ea0ba39494ce83a213fffba74279579268"), new(big.Int), 0, new(big.Int), nil),
			types.NewTransaction(2, common.HexToAddress("0xb794f5ea0ba39494ce83a213fffba74279579268"), new(big.Int), 0, new(big.Int), nil),
			types.NewTransaction(3, common.HexToAddress("0xb794f5ea0ba39494ce83a213fffba74279579268"), new(big.Int), 0, new(big.Int), nil),
			types.NewTransaction(4, common.HexToAddress("0xb794f5ea0ba39494ce83a213fffba74279579268"), new(big.Int), 0, new(big.Int), nil),
		}

		hashes []common.Hash
	)

	fid0 := api.NewPendingTransactionFilter()

	time.Sleep(1 * time.Second)
	backend.txsFeed.Send(core.NewTxsEvent{Txs: transactions})

	timeout := time.Now().Add(1 * time.Second)
	for {
		results, err := api.GetFilterChanges(fid0)
		if err != nil {
			t.Fatalf("Unable to retrieve logs: %v", err)
		}

		h := results.([]common.Hash)
		hashes = append(hashes, h...)
		if len(hashes) >= len(transactions) {
			break
		}
		// check timeout
		if time.Now().After(timeout) {
			break
		}

		time.Sleep(100 * time.Millisecond)
	}

	if len(hashes) != len(transactions) {
		t.Errorf("invalid number of transactions, want %d transactions(s), got %d", len(transactions), len(hashes))
		return
	}
	for i := range hashes {
		if hashes[i] != transactions[i].Hash() {
			t.Errorf("hashes[%d] invalid, want %x, got %x", i, transactions[i].Hash(), hashes[i])
		}
	}
}

// TestLogFilterCreation test whether a given filter criteria makes sense.
// If not it must return an error.
func TestLogFilterCreation(t *testing.T) {
	var (
		backend = newTestBackend()
		api     = NewPublicFilterAPI(backend)

		testCases = []struct {
			crit    FilterCriteria
			success bool
		}{
			// defaults
			{FilterCriteria{}, true},
			// valid block number range
			{FilterCriteria{FromBlock: big.NewInt(1), ToBlock: big.NewInt(2)}, true},
			// "mined" block range to pending
			{FilterCriteria{FromBlock: big.NewInt(1), ToBlock: big.NewInt(rpc.LatestBlockNumber.Int64())}, true},
			// new mined and pending blocks
			{FilterCriteria{FromBlock: big.NewInt(rpc.LatestBlockNumber.Int64()), ToBlock: big.NewInt(rpc.PendingBlockNumber.Int64())}, true},
			// from block "higher" than to block
			{FilterCriteria{FromBlock: big.NewInt(2), ToBlock: big.NewInt(1)}, false},
			// from block "higher" than to block
			{FilterCriteria{FromBlock: big.NewInt(rpc.LatestBlockNumber.Int64()), ToBlock: big.NewInt(100)}, false},
			// from block "higher" than to block
			{FilterCriteria{FromBlock: big.NewInt(rpc.PendingBlockNumber.Int64()), ToBlock: big.NewInt(100)}, false},
			// from block "higher" than to block
			{FilterCriteria{FromBlock: big.NewInt(rpc.PendingBlockNumber.Int64()), ToBlock: big.NewInt(rpc.LatestBlockNumber.Int64())}, false},
		}
	)

	for i, test := range testCases {
		_, err := api.NewFilter(test.crit)
		if test.success && err != nil {
			t.Errorf("expected filter creation for case %d to success, got %v", i, err)
		}
		if !test.success && err == nil {
			t.Errorf("expected testcase %d to fail with an error", i)
		}
	}
}

// TestInvalidLogFilterCreation tests whether invalid filter log criteria results in an error
// when the filter is created.
func TestInvalidLogFilterCreation(t *testing.T) {
	t.Parallel()

	var (
		backend = newTestBackend()
		api     = NewPublicFilterAPI(backend)
	)

	// different situations where log filter creation should fail.
	// Reason: fromBlock > toBlock
	testCases := []FilterCriteria{
		0: {FromBlock: big.NewInt(rpc.PendingBlockNumber.Int64()), ToBlock: big.NewInt(rpc.LatestBlockNumber.Int64())},
		1: {FromBlock: big.NewInt(rpc.PendingBlockNumber.Int64()), ToBlock: big.NewInt(100)},
		2: {FromBlock: big.NewInt(rpc.LatestBlockNumber.Int64()), ToBlock: big.NewInt(100)},
	}

	for i, test := range testCases {
		if _, err := api.NewFilter(test); err == nil {
			t.Errorf("Expected NewFilter for case #%d to fail", i)
		}
	}
}

func TestInvalidGetLogsRequest(t *testing.T) {
	var (
		backend   = newTestBackend()
		api       = NewPublicFilterAPI(backend)
		blockHash = common.HexToHash("0x1111111111111111111111111111111111111111111111111111111111111111")
	)

	// Reason: Cannot specify both BlockHash and FromBlock/ToBlock)
	testCases := []FilterCriteria{
		0: {BlockHash: &blockHash, FromBlock: big.NewInt(100)},
		1: {BlockHash: &blockHash, ToBlock: big.NewInt(500)},
		2: {BlockHash: &blockHash, FromBlock: big.NewInt(rpc.LatestBlockNumber.Int64())},
	}

	for i, test := range testCases {
		if _, err := api.GetLogs(context.Background(), test); err == nil {
			t.Errorf("Expected Logs for case #%d to fail", i)
		}
	}
}

// TestLogFilter tests whether log filters match the correct logs that are posted to the event feed.
func TestLogFilter(t *testing.T) {
	t.Parallel()

	var (
		backend = newTestBackend()
		api     = NewPublicFilterAPI(backend)

		firstAddr      = common.HexToAddress("0x1111111111111111111111111111111111111111")
		secondAddr     = common.HexToAddress("0x2222222222222222222222222222222222222222")
		thirdAddress   = common.HexToAddress("0x3333333333333333333333333333333333333333")
		notUsedAddress = common.HexToAddress("0x9999999999999999999999999999999999999999")
		firstTopic     = common.HexToHash("0x1111111111111111111111111111111111111111111111111111111111111111")
		secondTopic    = common.HexToHash("0x2222222222222222222222222222222222222222222222222222222222222222")
		notUsedTopic   = common.HexToHash("0x9999999999999999999999999999999999999999999999999999999999999999")

		allLogs = []*types.Log{
			{Address: firstAddr},
			{Address: firstAddr, Topics: []common.Hash{firstTopic}, BlockNumber: 1},
			{Address: secondAddr, Topics: []common.Hash{firstTopic}, BlockNumber: 1},
			{Address: thirdAddress, Topics: []common.Hash{secondTopic}, BlockNumber: 2},
			{Address: thirdAddress, Topics: []common.Hash{secondTopic}, BlockNumber: 3},
		}

		testCases = []struct {
			crit     FilterCriteria
			expected []*types.Log
			id       rpc.ID
		}{
			// match all
			0: {FilterCriteria{}, allLogs, ""},
			// match none due to no matching addresses
			1: {FilterCriteria{Addresses: []common.Address{{}, notUsedAddress}, Topics: [][]common.Hash{nil}}, []*types.Log{}, ""},
			// match logs based on addresses, ignore topics
			2: {FilterCriteria{Addresses: []common.Address{firstAddr}}, allLogs[:2], ""},
			// match none due to no matching topics (match with address)
			3: {FilterCriteria{Addresses: []common.Address{secondAddr}, Topics: [][]common.Hash{{notUsedTopic}}}, []*types.Log{}, ""},
			// match logs based on addresses and topics
			4: {FilterCriteria{Addresses: []common.Address{thirdAddress}, Topics: [][]common.Hash{{firstTopic, secondTopic}}}, allLogs[3:5], ""},
			// match logs based on multiple addresses and "or" topics
			5: {FilterCriteria{Addresses: []common.Address{secondAddr, thirdAddress}, Topics: [][]common.Hash{{firstTopic, secondTopic}}}, allLogs[2:5], ""},
			// all "mined" logs with block num >= 2
			6: {FilterCriteria{FromBlock: big.NewInt(2), ToBlock: big.NewInt(rpc.LatestBlockNumber.Int64())}, allLogs[3:], ""},
			// all "mined" logs
			7: {FilterCriteria{ToBlock: big.NewInt(rpc.LatestBlockNumber.Int64())}, allLogs, ""},
			// all "mined" logs with 1>= block num <=2 and topic secondTopic
			8: {FilterCriteria{FromBlock: big.NewInt(1), ToBlock: big.NewInt(2), Topics: [][]common.Hash{{secondTopic}}}, allLogs[3:4], ""},
			// match all logs due to wildcard topic
			9: {FilterCriteria{Topics: [][]common.Hash{nil}}, allLogs[1:], ""},
		}
	)

	// create all filters
	for i := range testCases {
		testCases[i].id, _ = api.NewFilter(testCases[i].crit)
	}

	// raise events
	time.Sleep(1 * time.Second)
	if nsend := backend.logsFeed.Send(allLogs); nsend == 0 {
		t.Fatal("Shoud have at least one subscription")
	}

	for i, tt := range testCases {
		var fetched []*types.Log
		timeout := time.Now().Add(1 * time.Second)
		for { // fetch all expected logs
			results, err := api.GetFilterChanges(tt.id)
			if err != nil {
				t.Fatalf("Unable to fetch logs: %v", err)
			}

			fetched = append(fetched, results.([]*types.Log)...)
			if len(fetched) >= len(tt.expected) {
				break
			}
			// check timeout
			if time.Now().After(timeout) {
				break
			}

			time.Sleep(100 * time.Millisecond)
		}

		if len(fetched) != len(tt.expected) {
			t.Errorf("invalid number of logs for case %d, want %d log(s), got %d", i, len(tt.expected), len(fetched))
			return
		}

		for l := range fetched {
			if fetched[l].Removed {
				t.Errorf("expected log not to be removed for log %d in case %d", l, i)
			}
			if !reflect.DeepEqual(fetched[l], tt.expected[l]) {
				t.Errorf("invalid log on index %d for case %d", l, i)
			}
		}
	}
}
