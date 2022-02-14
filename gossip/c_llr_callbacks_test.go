package gossip

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/Fantom-foundation/go-opera/eventcheck"
	"github.com/Fantom-foundation/go-opera/evmcore"
	"github.com/Fantom-foundation/go-opera/gossip/contract/ballot"
	"github.com/Fantom-foundation/go-opera/gossip/filters"
	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/inter/ibr"
	"github.com/Fantom-foundation/go-opera/inter/ier"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/inter/pos"
	"github.com/Fantom-foundation/lachesis-base/kvdb"

	"github.com/status-im/keycard-go/hexutils"
)

type IntegrationTestSuite struct {
	suite.Suite

	startEpoch, lastEpoch idx.Epoch
	generator, processor  *testEnv
	epochToEvsMap         map[idx.Epoch][]*inter.LlrSignedEpochVote
	bvs                   []*inter.LlrSignedBlockVotes
	blockIndices          []idx.Block
}

func (s *IntegrationTestSuite) SetupTest() {
	const (
		rounds        = 20
		validatorsNum = 10
		startEpoch    = 1
	)

	//creating generator and processor
	generator := newTestEnv(startEpoch, validatorsNum)
	processor := newTestEnv(startEpoch, validatorsNum)

	proposals := [][32]byte{
		ballotOption("Option 1"),
		ballotOption("Option 2"),
		ballotOption("Option 3"),
	}

	for n := uint64(0); n < rounds; n++ {
		txs := make([]*types.Transaction, validatorsNum)
		for i := idx.Validator(0); i < validatorsNum; i++ {
			_, tx, cBallot, err := ballot.DeployBallot(generator.Pay(idx.ValidatorID(i+1)), generator, proposals)
			s.Require().NoError(err)
			s.Require().NotNil(cBallot)
			s.Require().NotNil(tx)
			txs[i] = tx
		}
		tm := sameEpoch
		if n%10 == 0 {
			tm = nextEpoch
		}

		rr, err := generator.ApplyTxs(tm, txs...)
		s.Require().NoError(err)
		for _, r := range rr {
			s.Require().Len(r.Logs, 3)
			for _, l := range r.Logs {
				s.Require().NotNil(l)
			}
		}
	}

	s.startEpoch = startEpoch
	s.generator = generator
	s.processor = processor

	s.epochToEvsMap = fetchEvs(s.generator)
	bvs, blockIndices := fetchBvsBlockIdxs(s.generator)
	s.bvs = bvs
	s.blockIndices = blockIndices
	s.lastEpoch = generator.store.GetEpoch()
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down test suite")
	s.generator.Close()
	s.processor.Close()
}

// fetchEvs fetches LlrSignedEpochVotes from generator
func fetchEvs(generator *testEnv) map[idx.Epoch][]*inter.LlrSignedEpochVote {
	m := make(map[idx.Epoch][]*inter.LlrSignedEpochVote)
	it := generator.store.table.LlrEpochVotes.NewIterator(nil, nil)
	defer it.Release()
	for it.Next() {
		ev := &inter.LlrSignedEpochVote{}
		if err := rlp.DecodeBytes(it.Value(), ev); err != nil {
			generator.store.Log.Crit("Failed to decode epoch vote", "err", err)
		}

		if ev != nil {
			m[ev.Val.Epoch] = append(m[ev.Val.Epoch], ev)
		}
	}
	return m
}

// fetchBvsBlockIdxs fetches block indices of blocks that have min 4 LLR votes.
func fetchBvsBlockIdxs(generator *testEnv) ([]*inter.LlrSignedBlockVotes, []idx.Block) {
	var bvs []*inter.LlrSignedBlockVotes
	blockIdxCountMap := make(map[idx.Block]uint64)

	// fetching blockIndices with at least minVoteCount
	fetchBlockIdxs := func(blockIdxCountMap map[idx.Block]uint64) (blockIndices []idx.Block) {
		const minVoteCount = 4
		for blockIdx, count := range blockIdxCountMap {
			if count >= minVoteCount {
				blockIndices = append(blockIndices, blockIdx)
			}
		}
		return
	}

	// compute how any votes have been given for a particular block idx
	fillblockIdxCountMap := func(bv *inter.LlrSignedBlockVotes) {
		start, end := bv.Val.Start, bv.Val.Start+idx.Block(len(bv.Val.Votes))-1

		for b := start; start != 0 && b <= end; b++ {
			blockIdxCountMap[b] += 1
		}
	}

	it := generator.store.table.LlrBlockVotes.NewIterator(nil, nil)
	defer it.Release()
	for it.Next() {
		bv := &inter.LlrSignedBlockVotes{}
		if err := rlp.DecodeBytes(it.Value(), bv); err != nil {
			generator.store.Log.Crit("Failed to decode block vote while running ", "err", err)
		}

		if bv != nil {
			fillblockIdxCountMap(bv)
			bvs = append(bvs, bv)
		}
	}

	return bvs, fetchBlockIdxs(blockIdxCountMap)
}

// txByBlockSubsetOf iterates over procMap keys and checks equality of transaction hashes of genenerator and processor
func txByBlockSubsetOf(t *testing.T, procMap, genMap map[idx.Block]types.Transactions) {
	// assert len(procBlockToTxsMap.keys()) <= len(genBlockToTxsMap.keys()")
	require.LessOrEqual(t, len(reflect.ValueOf(procMap).MapKeys()), len(reflect.ValueOf(genMap).MapKeys()), "The number of keys does not match")
	for b, txs := range procMap {
		genTxs, ok := genMap[b]
		require.True(t, ok)
		require.Equal(t, len(txs), len(genTxs))
		for i, tx := range txs {
			require.Equal(t, tx.Hash().Hex(), genTxs[i].Hash().Hex())
		}
	}
}

// checkLogsEquality checks equality of logs slice field byf ield
func checkLogsEquality(t *testing.T, genLogs, procLogs []*types.Log) {
	require.Equal(t, len(genLogs), len(procLogs))
	for i, procLog := range procLogs {
		// compare all fields
		require.Equal(t, procLog.Address.Hex(), genLogs[i].Address.Hex())
		require.Equal(t, procLog.BlockHash.Hex(), genLogs[i].BlockHash.Hex())
		require.Equal(t, procLog.BlockNumber, genLogs[i].BlockNumber)
		require.Equal(t, hexutils.BytesToHex(procLog.Data), hexutils.BytesToHex(genLogs[i].Data))
		require.Equal(t, procLog.Index, genLogs[i].Index)
		require.Equal(t, procLog.Removed, genLogs[i].Removed)

		for j, topic := range procLog.Topics {
			require.Equal(t, topic.Hex(), genLogs[i].Topics[j].Hex())
		}

		require.Equal(t, procLog.TxHash.Hex(), genLogs[i].TxHash.Hex())
		require.Equal(t, procLog.TxIndex, genLogs[i].TxIndex)
	}
}

type testParams struct {
	t            *testing.T
	genEvmBlock  *evmcore.EvmBlock
	procEvmBlock *evmcore.EvmBlock
	genReceipts  types.Receipts
	procReceipts types.Receipts
}

func newTestParams(t *testing.T, genEvmBlock, procEvmBlock *evmcore.EvmBlock, genReceipts, procReceipts types.Receipts) testParams {
	return testParams{t, genEvmBlock, procEvmBlock, genReceipts, procReceipts}
}

func (p testParams) compareEvmBlocks() {
	// comparing all fields of genEvmBlock and procEvmBlock
	require.Equal(p.t, p.genEvmBlock.Number, p.procEvmBlock.Number)
	require.Equal(p.t, p.genEvmBlock.Hash, p.procEvmBlock.Hash)
	require.Equal(p.t, p.genEvmBlock.ParentHash, p.procEvmBlock.ParentHash)
	require.Equal(p.t, p.genEvmBlock.Root, p.procEvmBlock.Root)
	require.Equal(p.t, p.genEvmBlock.TxHash, p.procEvmBlock.TxHash)
	require.Equal(p.t, p.genEvmBlock.Time, p.procEvmBlock.Time)
	require.Equal(p.t, p.genEvmBlock.GasLimit, p.procEvmBlock.GasLimit)
	require.Equal(p.t, p.genEvmBlock.GasUsed, p.procEvmBlock.GasUsed)
	require.Equal(p.t, p.genEvmBlock.BaseFee, p.procEvmBlock.BaseFee)
}

func (p testParams) compareReceipts() {
	require.Equal(p.t, len(p.genReceipts), len(p.procReceipts))
	// compare every field except logs, I compare them separately
	for i, initRec := range p.genReceipts {
		require.Equal(p.t, initRec.Type, p.procReceipts[i].Type)
		require.Equal(p.t, hexutils.BytesToHex(initRec.PostState), hexutils.BytesToHex(p.procReceipts[i].PostState))
		require.Equal(p.t, initRec.Status, p.procReceipts[i].Status)
		require.Equal(p.t, initRec.CumulativeGasUsed, p.procReceipts[i].CumulativeGasUsed)
		require.Equal(p.t, hexutils.BytesToHex(initRec.Bloom.Bytes()), hexutils.BytesToHex(p.procReceipts[i].Bloom.Bytes()))
		require.Equal(p.t, initRec.TxHash.Hex(), p.procReceipts[i].TxHash.Hex())
		require.Equal(p.t, initRec.ContractAddress.Hex(), p.procReceipts[i].ContractAddress.Hex())
		require.Equal(p.t, initRec.GasUsed, p.procReceipts[i].GasUsed)
		require.Equal(p.t, initRec.BlockHash.String(), p.procReceipts[i].BlockHash.String())
		require.Equal(p.t, initRec.BlockNumber, p.procReceipts[i].BlockNumber)
		require.Equal(p.t, initRec.TransactionIndex, p.procReceipts[i].TransactionIndex)
	}
}

func (p testParams) compareLogs(initLogs2D, procLogs2D [][]*types.Log) {
	require.Equal(p.t, len(initLogs2D), len(procLogs2D))
	for i, initLogs := range initLogs2D {
		checkLogsEquality(p.t, initLogs, procLogs2D[i])
	}
}

func (p testParams) serializeAndCompare(val1, val2 interface{}) {
	// serialize val1 and val2
	buf1, err := rlp.EncodeToBytes(val1)
	require.NotNil(p.t, buf1)
	require.NoError(p.t, err)
	buf2, err := rlp.EncodeToBytes(val2)
	require.NotNil(p.t, buf2)
	require.NoError(p.t, err)

	// compare serialized representation of val1 and val2
	require.Equal(p.t, hexutils.BytesToHex(buf1), hexutils.BytesToHex(buf2))
}

func (p testParams) compareTransactions(initiator, processor *testEnv) {
	ctx := context.Background()
	require.Equal(p.t, len(p.genEvmBlock.Transactions), len(p.procEvmBlock.Transactions))
	for i, tx := range p.genEvmBlock.Transactions {
		txHash := tx.Hash()
		initTx, _, _, err := initiator.EthAPI.GetTransaction(ctx, txHash)
		require.NoError(p.t, err)

		procTx, _, _, err := processor.EthAPI.GetTransaction(ctx, txHash)
		require.NoError(p.t, err)

		require.Equal(p.t, txHash.Hex(), p.procEvmBlock.Transactions[i].Hash().Hex())
		require.Equal(p.t, txHash.Hex(), initTx.Hash().Hex())
		require.Equal(p.t, txHash.Hex(), procTx.Hash().Hex())
	}
}

func fetchTxsbyBlock(env *testEnv) map[idx.Block]types.Transactions {
	m := make(map[idx.Block]types.Transactions)
	it := env.store.table.Blocks.NewIterator(nil, nil)
	defer it.Release()
	for it.Next() {
		block := &inter.Block{}
		if err := rlp.DecodeBytes(it.Value(), block); err != nil {
			env.store.Log.Crit("Failed to decode block", "err", err)
		}

		if block != nil {
			n := idx.BytesToBlock(it.Key())
			txs := env.store.GetBlockTxs(n, block)
			m[n] = txs
		}
	}
	return m
}

type repeater struct {
	generator     *testEnv
	processor     *testEnv
	bvs           []*inter.LlrSignedBlockVotes
	blockIndices  []idx.Block
	epochToEvsMap map[idx.Epoch][]*inter.LlrSignedEpochVote
	t             *testing.T
}

func newRepeater(s *IntegrationTestSuite) repeater {
	return repeater{
		generator:     s.generator,
		processor:     s.processor,
		bvs:           s.bvs,
		blockIndices:  s.blockIndices,
		epochToEvsMap: s.epochToEvsMap,
		t:             s.T(),
	}
}

// processBlockVotesRecords processes block votes. Moreover, it processes block records for every block index that has minimum 4 LLr Votes.
// If ProcessFullBlockRecord returns an error, omit it in fullRepeater scenario, but not in testRepeater scenario.
func (r repeater) processBlockVotesRecords(isTestRepeater bool) {
	for _, bv := range r.bvs {
		r.processor.ProcessBlockVotes(*bv)
	}

	for _, blockIdx := range r.blockIndices {
		if br := r.generator.store.GetFullBlockRecord(blockIdx); br != nil {
			ibr := ibr.LlrIdxFullBlockRecord{LlrFullBlockRecord: *br, Idx: blockIdx}
			err := r.processor.ProcessFullBlockRecord(ibr)
			if err == nil {
				continue
			}

			// do not ingore this error in testRepeater
			if isTestRepeater {
				require.NoError(r.t, err)
			} else {
				// omit this error in fullRepeater
				require.EqualError(r.t, err, eventcheck.ErrAlreadyProcessedBR.Error())
			}

		} else {
			r.generator.Log.Crit("Empty full block record popped up")
		}
	}
}

// processEpochVotesRecords processes each epoch vote. Additionally, it processes epoch block records in range [startEpoch+1; lastEpoch]
func (r repeater) processEpochVotesRecords(startEpoch, lastEpoch idx.Epoch) {
	// invoke repeater.ProcessEpochVote and ProcessFullEpochRecord for epoch in range [2; lastepoch]
	for e := idx.Epoch(startEpoch + 1); e <= lastEpoch; e++ {
		epochVotes, ok := r.epochToEvsMap[e]
		if !ok {
			r.processor.store.Log.Crit("Failed to fetch epoch votes for a given epoch")
		}

		for _, v := range epochVotes {
			require.NoError(r.t, r.processor.ProcessEpochVote(*v))
		}

		if er := r.generator.store.GetFullEpochRecord(e); er != nil {
			ier := ier.LlrIdxFullEpochRecord{LlrFullEpochRecord: *er, Idx: e}
			require.NoError(r.t, r.processor.ProcessFullEpochRecord(ier))
		}
	}
}

// compareERHashes compares epoch recors hashes. Moreover, it checks equality of hashes of epoch and block states.
func (r repeater) compareERHashes(startEpoch, lastEpoch idx.Epoch) {
	for e := startEpoch; e <= lastEpoch; e++ {

		genBs, genEs := r.generator.store.GetHistoryBlockEpochState(e)
		repBs, repEs := r.processor.store.GetHistoryBlockEpochState(e)
		require.Equal(r.t, genBs.Hash().Hex(), repBs.Hash().Hex())
		require.Equal(r.t, genEs.Hash().Hex(), repEs.Hash().Hex())

		genEr := r.generator.store.GetFullEpochRecord(e)
		repEr := r.processor.store.GetFullEpochRecord(e)
		require.Equal(r.t, genEr.Hash().Hex(), repEr.Hash().Hex())
	}
}

// compareParams checks equality of different parameters such as BlockByHash, BlockByNumber, Receipts, Logs
func (r repeater) compareParams() {
	ctx := context.Background()

	// compare blockbyNumber
	for _, blockIdx := range r.blockIndices {

		// comparing EvmBlock by calling BlockByHash
		genEvmBlock, err := r.generator.EthAPI.BlockByNumber(ctx, rpc.BlockNumber(blockIdx))
		require.NotNil(r.t, genEvmBlock)
		require.NoError(r.t, err)

		procEvmBlock, err := r.processor.EthAPI.BlockByNumber(ctx, rpc.BlockNumber(blockIdx))
		require.NotNil(r.t, procEvmBlock)
		require.NoError(r.t, err)

		// compare Receipts
		genReceipts := r.generator.store.evm.GetReceipts(blockIdx, r.generator.EthAPI.signer, genEvmBlock.Hash, genEvmBlock.Transactions)
		require.NotNil(r.t, genReceipts)
		procReceipts := r.processor.store.evm.GetReceipts(blockIdx, r.processor.EthAPI.signer, procEvmBlock.Hash, procEvmBlock.Transactions)
		require.NotNil(r.t, procReceipts)

		testParams := newTestParams(r.t, genEvmBlock, procEvmBlock, genReceipts, procReceipts)
		testParams.compareEvmBlocks()

		r.t.Log("comparing receipts")
		testParams.compareReceipts()

		// comparing evmBlock by calling BlockByHash
		genEvmBlock, err = r.generator.EthAPI.BlockByHash(ctx, genEvmBlock.Hash)
		require.NotNil(r.t, genEvmBlock)
		require.NoError(r.t, err)
		procEvmBlock, err = r.processor.EthAPI.BlockByHash(ctx, procEvmBlock.Hash)
		require.NotNil(r.t, procEvmBlock)
		require.NoError(r.t, err)

		testParams = newTestParams(r.t, genEvmBlock, procEvmBlock, genReceipts, procReceipts)
		testParams.compareEvmBlocks()

		// compare Logs
		genLogs, err := r.generator.EthAPI.GetLogs(ctx, genEvmBlock.Hash)
		require.NoError(r.t, err)

		procLogs, err := r.processor.EthAPI.GetLogs(ctx, genEvmBlock.Hash)
		require.NoError(r.t, err)

		r.t.Log("comparing logs")
		testParams.serializeAndCompare(genLogs, procLogs)
		testParams.compareLogs(genLogs, procLogs)

		// compare ReceiptForStorage
		genBR := r.generator.store.GetFullBlockRecord(blockIdx)
		procBR := r.processor.store.GetFullBlockRecord(blockIdx)
		testParams.serializeAndCompare(genBR.Receipts, procBR.Receipts)

		// compare BR hashes
		require.Equal(r.t, genBR.Hash().Hex(), procBR.Hash().Hex())

		// compare transactions
		testParams.compareTransactions(r.generator, r.processor)
	}
}

// compareLogsByFilterCriteria introduces testing logic for GetLogs function for generator and processor
func (r repeater) compareLogsByFilterCriteria() {
	var crit filters.FilterCriteria

	blockIdxLogsMap := func() map[idx.Block][]*types.Log {
		ctx := context.Background()
		m := make(map[idx.Block][]*types.Log, len(r.blockIndices))

		for _, blockIdx := range r.blockIndices {
			block, err := r.generator.EthAPI.BlockByNumber(ctx, rpc.BlockNumber(blockIdx))
			require.NotNil(r.t, block)
			require.NoError(r.t, err)
			receipts := r.generator.store.evm.GetReceipts(blockIdx, r.generator.EthAPI.signer, block.Hash, block.Transactions)
			for _, r := range receipts {
				// we add only non empty logs
				if len(r.Logs) > 0 {
					m[blockIdx] = append(m[blockIdx], r.Logs...)
				}
			}

		}
		return m
	}()

	findLastNonEmptyLogs := func() (idx.Block, []*types.Log, error) {
		for i := len(r.blockIndices) - 1; i >= 0; i-- {
			logs, ok := blockIdxLogsMap[r.blockIndices[i]]
			if !ok {
				continue
			}
			if len(logs) > 0 {
				return r.blockIndices[i], logs, nil
			}
		}

		return 0, nil, errors.New("all blocks have no logs")
	}

	lastBlockNumber, lastLogs, err := findLastNonEmptyLogs()
	require.NoError(r.t, err)
	require.NotNil(r.t, lastLogs)

	defaultCrit := filters.FilterCriteria{FromBlock: big.NewInt(1), ToBlock: big.NewInt(int64(lastBlockNumber/2 + 1))}

	ctx := context.Background()

	config := filters.DefaultConfig()
	config.UnindexedLogsBlockRangeLimit = idx.Block(1000)
	genApi := filters.NewPublicFilterAPI(r.generator.EthAPI, config)
	require.NotNil(r.t, genApi)

	procApi := filters.NewPublicFilterAPI(r.processor.EthAPI, config)
	require.NotNil(r.t, procApi)

	defaultLogs, err := genApi.GetLogs(ctx, defaultCrit)
	require.NoError(r.t, err)
	require.NotNil(r.t, defaultLogs)
	require.NotEqual(r.t, defaultLogs, []*types.Log{})

	findFirstNonEmptyLogs := func() (idx.Block, []*types.Log, error) {
		for _, blockIdx := range r.blockIndices {
			logs, ok := blockIdxLogsMap[blockIdx]
			if !ok {
				continue
			}
			if len(logs) > 0 {
				return blockIdx, logs, nil
			}
		}

		return 0, nil, errors.New("all blocks have no logs")
	}

	fetchFirstAddrFromLogs := func(logs []*types.Log) (common.Address, error) {
		for i := range logs {
			if logs[i] != nil && logs[i].Address != (common.Address{}) {
				return logs[i].Address, nil
			}
		}

		return common.Address{}, errors.New("no address can be found in logs")
	}

	fetchFirstTopicFromLogs := func(logs []*types.Log) (common.Hash, error) {
		for i := range logs {
			if logs[i] != nil && logs[i].Topics[0] != (common.Hash{}) {
				return logs[i].Topics[0], nil
			}
		}
		return common.Hash{}, errors.New("no topic can be found in logs")
	}

	fetchRandomTopicFromLogs := func(logs []*types.Log) common.Hash {
		rand.Seed(time.Now().Unix())
		l := rand.Int() % len(logs) // pick log at random

		rand.Seed(time.Now().Unix())
		t := rand.Int() % len(logs[l].Topics) // pick topic at random

		return logs[l].Topics[t]
	}

	fetchRandomAddrFromLogs := func(logs []*types.Log) common.Address {
		rand.Seed(time.Now().Unix())
		l := rand.Int() % len(logs) // pick log at random

		return logs[l].Address
	}

	blockNumber, logs, err := findFirstNonEmptyLogs()
	require.NoError(r.t, err)
	require.NotNil(r.t, logs)

	firstAddr, err := fetchFirstAddrFromLogs(logs)
	require.NoError(r.t, err)

	firstTopic, err := fetchFirstTopicFromLogs(logs)
	require.NoError(r.t, err)

	lastAddr, err := fetchFirstAddrFromLogs(lastLogs)
	require.NoError(r.t, err)

	lastTopic, err := fetchFirstTopicFromLogs(lastLogs)
	require.NoError(r.t, err)

	testCases := []struct {
		name    string
		pretest func()
		success bool
	}{
		{"single valid address",
			func() {
				crit = filters.FilterCriteria{
					FromBlock: big.NewInt(int64(blockNumber)),
					ToBlock:   big.NewInt(int64(blockNumber)),
					Addresses: []common.Address{firstAddr},
				}
			},
			true,
		},
		{"single invalid address",
			func() {
				invalidAddr := common.BytesToAddress([]byte("invalid address"))
				crit = filters.FilterCriteria{
					FromBlock: big.NewInt(int64(blockNumber)),
					ToBlock:   big.NewInt(int64(blockNumber)),
					Addresses: []common.Address{invalidAddr},
				}
			},
			false,
		},
		{"invalid block range",
			func() {
				crit = filters.FilterCriteria{
					FromBlock: big.NewInt(int64(blockNumber) + 1),
					ToBlock:   big.NewInt(int64(blockNumber) + 2),
					Addresses: []common.Address{firstAddr},
				}
			},
			false,
		},
		{"block range 1-1000",
			func() {
				crit = defaultCrit
			},
			true,
		},
		{"block range 1-1000 and first topic",
			func() {
				crit = defaultCrit
				crit.Topics = [][]common.Hash{{firstTopic}}
			},
			true,
		},
		{"block range 1-1000 and random topic",
			func() {
				randomTopic := fetchRandomTopicFromLogs(defaultLogs)
				crit = defaultCrit
				crit.Topics = [][]common.Hash{{randomTopic}}
			},
			true,
		},
		{"block range 1-1000 and first address",
			func() {
				crit = defaultCrit
				crit.Addresses = []common.Address{firstAddr}
			},
			true,
		},
		{"block range 1-1000 and random address",
			func() {
				randomAddress := fetchRandomAddrFromLogs(defaultLogs)
				crit = defaultCrit
				crit.Addresses = []common.Address{randomAddress}
			},
			true,
		},
		{"block range 1 to lastBlockNumber",
			func() {
				crit = filters.FilterCriteria{
					FromBlock: big.NewInt(int64(1)),
					ToBlock:   big.NewInt(int64(lastBlockNumber)),
				}
			},
			true,
		},
		{"block range 1 to lastBlockNumber and last topic",
			func() {
				crit = filters.FilterCriteria{
					FromBlock: big.NewInt(int64(1)),
					ToBlock:   big.NewInt(int64(lastBlockNumber)),
					Topics:    [][]common.Hash{{lastTopic}},
				}
			},
			true,
		},
		{"block range 1 to lastBlockNumber, last address",
			func() {
				crit = filters.FilterCriteria{
					FromBlock: big.NewInt(int64(1)),
					ToBlock:   big.NewInt(int64(lastBlockNumber)),
					Addresses: []common.Address{lastAddr},
				}
			},
			true,
		},

		{"block range is nil and last address",
			func() {
				crit = filters.FilterCriteria{
					Addresses: []common.Address{lastAddr},
				}
			},
			true,
		},
		{"block range is nil and invalid address",
			func() {
				invalidAddr := common.BytesToAddress([]byte("invalid addr"))
				crit = filters.FilterCriteria{
					Addresses: []common.Address{invalidAddr},
				}
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		r.t.Run(tc.name, func(t *testing.T) {
			tc.pretest()
			genLogs, genErr := genApi.GetLogs(ctx, crit)
			procLogs, procErr := procApi.GetLogs(ctx, crit)
			if tc.success {
				require.NoError(t, procErr)
				require.NoError(t, genErr)
				checkLogsEquality(t, genLogs, procLogs)
			} else {
				require.Equal(t, genLogs, []*types.Log{})
				require.Equal(t, procLogs, []*types.Log{})
			}
		})
	}

	// iterative tests with random address and random topic
	itTestCases := []struct {
		name    string
		rounds  int
		pretest func()
	}{
		{"block range 1-1000 and random topic",
			100,
			func() {
				randomTopic := fetchRandomTopicFromLogs(defaultLogs)
				crit = defaultCrit
				crit.Topics = [][]common.Hash{{randomTopic}}
			},
		},
		{"block range 1-1000 and random address",
			100,
			func() {
				randomAddress := fetchRandomAddrFromLogs(defaultLogs)
				crit = defaultCrit
				crit.Addresses = []common.Address{randomAddress}
			},
		},
	}

	for _, tc := range itTestCases {
		tc := tc
		r.t.Run(tc.name, func(t *testing.T) {
			for i := 0; i < tc.rounds; i++ {
				tc.pretest()
				genLogs, genErr := genApi.GetLogs(ctx, crit)
				procLogs, procErr := procApi.GetLogs(ctx, crit)
				require.NoError(t, procErr)
				require.NoError(t, genErr)
				checkLogsEquality(t, genLogs, procLogs)
			}
		})
	}
}

// use command `go test  -timeout 120s  -run ^TestIntegrationTestSuite$ -testify.m TestRepeater` to run test scenario
func (s *IntegrationTestSuite) TestRepeater() {
	repeater := newRepeater(s)
	repeater.processEpochVotesRecords(s.startEpoch, s.lastEpoch)
	repeater.processBlockVotesRecords(true)

	s.Require().NoError(s.generator.store.Commit())
	s.Require().NoError(s.processor.store.Commit())

	// Compare transaction hashes
	s.T().Log("Checking procBlockToTxsMap <= genBlockToTxsMap")
	genBlockToTxsMap := fetchTxsbyBlock(s.generator)
	procBlockToTxsMap := fetchTxsbyBlock(s.processor)
	txByBlockSubsetOf(s.T(), procBlockToTxsMap, genBlockToTxsMap)

	// compare ER hashes
	repeater.compareERHashes(s.startEpoch+1, s.lastEpoch)
	// compare different parameters such as Logs,Receipts, Blockhash etc
	repeater.compareParams()

	// compare Logs by different criteria
	repeater.compareLogsByFilterCriteria()
}

// use command `go test  -timeout 120s  -run ^TestIntegrationTestSuite$ -testify.m TestFullRepeater` to run test scenario
func (s *IntegrationTestSuite) TestFullRepeater() {

	fullRepeater := newRepeater(s)

	wg := new(sync.WaitGroup)
	wg.Add(2)
	go func() {
		defer wg.Done()
		// process LLR epochVotes in fullRepeater
		fullRepeater.processEpochVotesRecords(s.startEpoch, s.lastEpoch)

		// process LLR block votes and BRs in fullReapeter
		fullRepeater.processBlockVotesRecords(false)

	}()

	go func() {
		defer wg.Done()
		events := func() (events []*inter.EventPayload) {
			it := s.generator.store.table.Events.NewIterator(nil, nil)
			defer it.Release()
			for it.Next() {
				e := &inter.EventPayload{}
				if err := rlp.DecodeBytes(it.Value(), e); err != nil {
					s.generator.store.Log.Crit("Failed to decode event", "err", err)
				}
				if e != nil {
					events = append(events, e)
				}
			}
			return
		}()

		for _, e := range events {
			s.processor.engineMu.Lock()
			s.Require().NoError(s.processor.processEvent(e))
			s.processor.engineMu.Unlock()
		}
	}()

	wg.Wait()

	// Comparing the store states
	fetchTable := func(table kvdb.Store) map[string]string {
		var m = make(map[string]string)
		it := table.NewIterator(nil, nil)
		defer it.Release()
		for it.Next() {
			key, value := it.Key(), it.Value()
			m[string(key)] = string(value)
		}
		return m
	}

	s.Require().NoError(s.generator.store.Commit())
	s.Require().NoError(s.processor.store.Commit())

	// Comparing generator and fullRepeater states

	// 1.Comparing Tx hashes
	s.T().Log("Checking genBlockToTxsMap <= procRepBlockToTxsMap")
	genBlockToTxsMap := fetchTxsbyBlock(s.generator)
	procBlockToTxsMap := fetchTxsbyBlock(s.processor)
	txByBlockSubsetOf(s.T(), procBlockToTxsMap, genBlockToTxsMap)

	// 2.Compare BlockByNumber,BlockByhash, GetReceipts, GetLogs
	fullRepeater.compareParams()

	// 2. Comparing mainDb of generator and fullRepeater
	genKVMap := fetchTable(s.generator.store.mainDB)
	fullRepKVMap := fetchTable(s.processor.store.mainDB)

	subsetOf := func(aa, bb map[string]string) {
		for _k, _v := range aa {
			k, v := []byte(_k), []byte(_v)
			if k[0] == 0 || k[0] == 'x' || k[0] == 'X' || k[0] == 'b' || k[0] == 'S' {
				continue
			}
			s.Require().Equal(hexutils.BytesToHex(v), hexutils.BytesToHex([]byte(bb[_k])))
		}
	}

	checkEqual := func(aa, bb map[string]string) {
		subsetOf(aa, bb)
		subsetOf(bb, aa)
	}

	s.T().Log("Checking genKVs == fullKVs")
	checkEqual(genKVMap, fullRepKVMap)

	genKVMapAfterIndexLogs := fetchTable(s.generator.store.mainDB)
	fullRepKVMapAfterIndexLogs := fetchTable(s.processor.store.mainDB)

	// comparing the states
	checkEqual(genKVMap, genKVMapAfterIndexLogs)
	checkEqual(fullRepKVMap, fullRepKVMapAfterIndexLogs)
	checkEqual(genKVMapAfterIndexLogs, fullRepKVMapAfterIndexLogs)

	fullRepeater.compareLogsByFilterCriteria()
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func TestBlockAndEpochRecords(t *testing.T) {
	const (
		validatorsNum = 10
		startEpoch    = 1
	)
	// setup testEnv
	env := newTestEnv(startEpoch, validatorsNum)

	// 1.create epoch record er1 manually
	er1 := ier.LlrIdxFullEpochRecord{Idx: idx.Epoch(startEpoch) + 1}
	er1Hash := er1.Hash()
	// 3. process ER1, the error will be popped up.
	require.EqualError(t, env.ProcessFullEpochRecord(er1), eventcheck.ErrUndecidedER.Error())

	// 2.create block record manually
	br1 := ibr.LlrIdxFullBlockRecord{Idx: idx.Block(2)}
	br1Hash := br1.Hash()
	//3. process BR1, the error will popped up
	require.EqualError(t, env.ProcessFullBlockRecord(br1), eventcheck.ErrUndecidedBR.Error())

	// 4.create less than 1/3W+1 epoch votes, ER1 still should not be processed
	for i := 1; i < 4; i++ {
		e := fakeEvent(0, 0, 0, 2, i, er1Hash, true)
		ev := inter.AsSignedEpochVote(e)
		require.NoError(t, env.ProcessEpochVote(ev))
	}
	require.EqualError(t, env.ProcessFullEpochRecord(er1), eventcheck.ErrUndecidedER.Error())

	// 5. add one more epoch vote,so 4 = 1/3W+1. Hence, ER1 has to be processed.
	fmt.Println("adding 4th epoch vote")
	e := fakeEvent(0, 0, 0, 2, 4, er1Hash, true)
	ev := inter.AsSignedEpochVote(e)
	require.NoError(t, env.ProcessEpochVote(ev))
	require.NoError(t, env.ProcessFullEpochRecord(er1))

	// 6.create epoch record er2  of same epoch as er1, but with another name.
	er2 := ier.LlrIdxFullEpochRecord{Idx: idx.Epoch(startEpoch + 1)}
	// 7.Get an error that the er has been already processed.
	require.EqualError(t, env.ProcessFullEpochRecord(er2), eventcheck.ErrAlreadyProcessedER.Error())

	//8. try to process Br1 with one vote with the same epoch as er1. it will(*Validators).GetWeightByIdx(...)
	e = fakeEvent(0, 1, er1.Idx, 2, 0, br1Hash, false)
	bv := inter.AsSignedBlockVotes(e)
	require.Panics(t, func() { env.ProcessBlockVotes(bv) }) //cause there are no validators
	require.EqualError(t, env.ProcessFullBlockRecord(br1), eventcheck.ErrUndecidedBR.Error())

	//9,10. process er1 and er2. it should yield an ErrAlreadyProcessedER error
	require.EqualError(t, env.ProcessFullEpochRecord(er1), eventcheck.ErrAlreadyProcessedER.Error())
	require.EqualError(t, env.ProcessFullEpochRecord(er2), eventcheck.ErrAlreadyProcessedER.Error())

	//11 add votes < 1/3W+1 for Br1. Record still should not be processed.
	fmt.Println("adding 3 votes for br1")
	for i := 5; i < 8; i++ {
		e := fakeEvent(0, 1, 0, 2, i, br1Hash, true)
		bv := inter.AsSignedBlockVotes(e)
		require.NoError(t, env.ProcessBlockVotes(bv))
	}
	require.EqualError(t, env.ProcessFullBlockRecord(br1), eventcheck.ErrUndecidedBR.Error())

	//12 add one vote for br1, then we have 1/3W+1 votes
	fmt.Println("adding 4th block vote to make up to match 1/3W+1")
	e = fakeEvent(0, 1, 0, 2, 8, br1Hash, true)
	bv = inter.AsSignedBlockVotes(e)
	require.NoError(t, env.ProcessBlockVotes(bv))

	// 13. create one more record of the same block, but different.
	br2 := ibr.LlrIdxFullBlockRecord{LlrFullBlockRecord: ibr.LlrFullBlockRecord{GasUsed: 100505}, Idx: idx.Block(2)}

	// 14. process br2. It should output an error, that block record hash is mismatched.
	require.EqualError(t, env.ProcessFullBlockRecord(br2), errors.New("block record hash mismatch").Error())

	// 15 process br1
	require.NoError(t, env.ProcessFullBlockRecord(br1))

	//16 process br1 and br2, they should yield an error that they have been already processed
	require.EqualError(t, env.ProcessFullBlockRecord(br1), eventcheck.ErrAlreadyProcessedBR.Error())
	require.EqualError(t, env.ProcessFullBlockRecord(br2), eventcheck.ErrAlreadyProcessedBR.Error())
}

// can not import it from inter package (((
func fakeEvent(txsNum, bvsNum int, bvEpoch, evEpoch idx.Epoch, valID int, recordHash hash.Hash, ersNum bool) *inter.EventPayload {
	r := rand.New(rand.NewSource(int64(0)))
	random := &inter.MutableEventPayload{}
	random.SetVersion(1)
	random.SetEpoch(2)
	random.SetNetForkID(uint16(r.Uint32() >> 16))
	random.SetLamport(idx.Lamport(rand.Intn(100) + 900))
	random.SetExtra([]byte{byte(r.Uint32())})
	random.SetSeq(idx.Event(r.Uint32() >> 8))
	random.SetCreator(idx.ValidatorID(valID))
	random.SetFrame(idx.Frame(r.Uint32() >> 16))
	random.SetCreationTime(inter.Timestamp(r.Uint64()))
	random.SetMedianTime(inter.Timestamp(r.Uint64()))
	random.SetGasPowerUsed(r.Uint64())
	random.SetGasPowerLeft(inter.GasPowerLeft{[2]uint64{r.Uint64(), r.Uint64()}})
	txs := types.Transactions{}
	for i := 0; i < txsNum; i++ {
		h := hash.Hash{}
		r.Read(h[:])
		if i%3 == 0 {
			tx := types.NewTx(&types.LegacyTx{
				Nonce:    r.Uint64(),
				GasPrice: randBig(r),
				Gas:      257 + r.Uint64(),
				To:       nil,
				Value:    randBig(r),
				Data:     randBytes(r, r.Intn(300)),
				V:        big.NewInt(int64(r.Intn(0xffffffff))),
				R:        h.Big(),
				S:        h.Big(),
			})
			txs = append(txs, tx)
		} else if i%3 == 1 {
			tx := types.NewTx(&types.AccessListTx{
				ChainID:    randBig(r),
				Nonce:      r.Uint64(),
				GasPrice:   randBig(r),
				Gas:        r.Uint64(),
				To:         randAddrPtr(r),
				Value:      randBig(r),
				Data:       randBytes(r, r.Intn(300)),
				AccessList: randAccessList(r, 300, 300),
				V:          big.NewInt(int64(r.Intn(0xffffffff))),
				R:          h.Big(),
				S:          h.Big(),
			})
			txs = append(txs, tx)
		} else {
			tx := types.NewTx(&types.DynamicFeeTx{
				ChainID:    randBig(r),
				Nonce:      r.Uint64(),
				GasTipCap:  randBig(r),
				GasFeeCap:  randBig(r),
				Gas:        r.Uint64(),
				To:         randAddrPtr(r),
				Value:      randBig(r),
				Data:       randBytes(r, r.Intn(300)),
				AccessList: randAccessList(r, 300, 300),
				V:          big.NewInt(int64(r.Intn(0xffffffff))),
				R:          h.Big(),
				S:          h.Big(),
			})
			txs = append(txs, tx)
		}
	}

	bvs := inter.LlrBlockVotes{}
	if bvsNum > 0 {
		bvs.Start = 2
		switch {
		case bvEpoch > 0:
			bvs.Epoch = bvEpoch
			random.SetEpoch(bvEpoch)
		default:
			bvs.Epoch = 1
			random.SetEpoch(1)
		}
	}

	for i := 0; i < bvsNum; i++ {
		bvs.Votes = append(bvs.Votes, recordHash)
	}

	ev := inter.LlrEpochVote{}
	if ersNum {
		ev.Epoch = evEpoch
		ev.Vote = recordHash
	}

	random.SetTxs(txs)
	random.SetEpochVote(ev)
	random.SetBlockVotes(bvs)
	random.SetPayloadHash(inter.CalcPayloadHash(random))

	parent := inter.MutableEventPayload{}
	parent.SetVersion(1)
	parent.SetLamport(random.Lamport() - 500)
	parent.SetEpoch(random.Epoch())
	random.SetParents(hash.Events{parent.Build().ID()})

	return random.Build()
}

func randBig(r *rand.Rand) *big.Int {
	b := make([]byte, r.Intn(8))
	_, _ = r.Read(b)
	if len(b) == 0 {
		b = []byte{0}
	}
	return new(big.Int).SetBytes(b)
}

func randBytes(r *rand.Rand, size int) []byte {
	b := make([]byte, size)
	r.Read(b)
	return b
}

func randAddrPtr(r *rand.Rand) *common.Address {
	addr := randAddr(r)
	return &addr
}

func randAddr(r *rand.Rand) common.Address {
	addr := common.Address{}
	r.Read(addr[:])
	return addr
}

func randAccessList(r *rand.Rand, maxAddrs, maxKeys int) types.AccessList {
	accessList := make(types.AccessList, r.Intn(maxAddrs))
	for i := range accessList {
		accessList[i].Address = randAddr(r)
		accessList[i].StorageKeys = make([]common.Hash, r.Intn(maxKeys))
		for j := range accessList[i].StorageKeys {
			r.Read(accessList[i].StorageKeys[j][:])
		}
	}
	return accessList
}

func TestEpochRecordWithDiffValidators(t *testing.T) {
	const (
		validatorsNum = 10
		startEpoch    = 2
	)
	require := require.New(t)
	// setup testEnv
	env := newTestEnv(startEpoch, validatorsNum)

	// Стартвые валидаторы имеют равномерные веса, стартовая эпоха - 2
	bs, es := env.store.GetHistoryBlockEpochState(startEpoch)

	// get new validators with different votes
	newVals := func() *pos.Validators {
		builder := pos.NewBuilder()
		defaultWeight := pos.Weight(111022302)
		for i := idx.ValidatorID(1); i <= 10; i++ {
			w := defaultWeight
			if i%2 == 0 {
				w -= 10021567
			} else {
				w += 10021567
			}
			builder.Set(i, w)
		}
		return builder.Build()
	}()

	// save new validators to state of epoch 2
	esCopy := es.Copy()
	esCopy.Validators = newVals

	// process ER of 3rd epoch
	er := ier.LlrIdxFullEpochRecord{
		LlrFullEpochRecord: ier.LlrFullEpochRecord{*bs, esCopy},
		Idx:                idx.Epoch(startEpoch + 1),
	}
	erHash := er.Hash()

	for i := 1; i <= 4; i++ {
		e := fakeEvent(0, 0, 0, startEpoch+1, i, erHash, true)
		ev := inter.AsSignedEpochVote(e)
		// process validators with equal weights
		require.NoError(env.ProcessEpochVote(ev))
	}

	require.NoError(env.ProcessFullEpochRecord(er))

	// process ER of 4th epoch with validators with different weights

	// get bs and es of 3rd apoch
	bs, es = env.store.GetHistoryBlockEpochState(startEpoch + 1)

	// put es and bs of 3rd apoch at LlrIdxFullEpochRecord of epoch 4
	er = ier.LlrIdxFullEpochRecord{
		LlrFullEpochRecord: ier.LlrFullEpochRecord{*bs, *es},
		Idx:                idx.Epoch(startEpoch + 2)}
	erHash = er.Hash()

	// confirm with votes of different weights
	for i := 1; i <= 5; i++ {
		e := fakeEvent(0, 0, 0, startEpoch+2, i, erHash, true)
		ev := inter.AsSignedEpochVote(e)
		require.NoError(env.ProcessEpochVote(ev))
	}

	// process ER of epoch 4
	require.NoError(env.ProcessFullEpochRecord(er))

	// process ER for epoch 5
	bs, es = env.store.GetHistoryBlockEpochState(startEpoch + 2)

	// yield validators with unequal weights to process in epoch 6
	newVals, partialWeight := func() (*pos.Validators, pos.Weight) {
		builder := pos.NewBuilder()
		w := pos.Weight(1000)

		// set 7 validators with weight 1000
		var partialWeight pos.Weight
		for i := idx.ValidatorID(1); i <= 7; i++ {
			partialWeight += w
			builder.Set(i, w)
		}

		w = pos.Weight(1000000)
		//set 8th, 9th and 10th validatora with weight 1000000
		for i := idx.ValidatorID(8); i <= 10; i++ {
			builder.Set(i, w)
		}

		return builder.Build(), partialWeight
	}()

	// save new validators to state
	esCopy = es.Copy()
	esCopy.Validators = newVals

	er = ier.LlrIdxFullEpochRecord{
		LlrFullEpochRecord: ier.LlrFullEpochRecord{*bs, esCopy},
		Idx:                idx.Epoch(startEpoch + 3),
	}
	erHash = er.Hash()

	for i := 1; i <= 10; i++ {
		e := fakeEvent(0, 0, 0, startEpoch+3, i, erHash, true)
		ev := inter.AsSignedEpochVote(e)
		require.NoError(env.ProcessEpochVote(ev))
	}

	require.NoError(env.ProcessFullEpochRecord(er))

	// process ER with epoch 6
	bs, es = env.store.GetHistoryBlockEpochState(startEpoch + 3)

	er = ier.LlrIdxFullEpochRecord{
		LlrFullEpochRecord: ier.LlrFullEpochRecord{*bs, *es},
		Idx:                idx.Epoch(startEpoch + 4),
	}
	erHash = er.Hash()

	for i := 1; i <= 7; i++ {
		e := fakeEvent(0, 0, 0, startEpoch+4, i, erHash, true)
		ev := inter.AsSignedEpochVote(e)
		require.NoError(env.ProcessEpochVote(ev))
	}

	// process ER for epoch 6
	// threshold weight is 1002334 = 1/3W+1
	// 7 validators with total weight 7000 is less than threshold weight
	// so 7 votes are not enough
	totalWeight := newVals.TotalWeight()
	thresholdWeight := pos.Weight(totalWeight/3 + 1)
	require.Less(partialWeight, thresholdWeight)
	require.EqualError(env.ProcessFullEpochRecord(er), eventcheck.ErrUndecidedER.Error())
}
