package gossip

import (
	"context"
	"errors"
	"math/big"
	"math/rand"
	"time"

	//"math/rand"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	//"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/Fantom-foundation/go-opera/eventcheck"
	"github.com/Fantom-foundation/go-opera/evmcore"
	"github.com/Fantom-foundation/go-opera/gossip/filters"
	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/inter/ibr"
	"github.com/Fantom-foundation/go-opera/inter/ier"

	"github.com/Fantom-foundation/go-opera/gossip/contract/ballot"

	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/kvdb"

	"github.com/status-im/keycard-go/hexutils"
)

type IntegrationTestSuite struct {
	suite.Suite

	startEpoch, lastEpoch idx.Epoch
	generator, processor  *testEnv
	epochToEvsMap map[idx.Epoch][]*inter.LlrSignedEpochVote
	bvs []*inter.LlrSignedBlockVotes
	blockIndices []idx.Block
	genBlockToTxsMap, procBlockToTxsMap map[idx.Block]types.Transactions

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
		// TODO grab logs only from specific blockhash for thosewere given more or equal 4 votes
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
	s.genBlockToTxsMap = fetchTxsbyBlock(generator)
	s.procBlockToTxsMap = fetchTxsbyBlock(processor)

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



// fetchBvsBlockIdxs computes block indices of blocks that have min 4 votes.
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




func txByBlockSubsetOf(t *testing.T, repMap, genMap map[idx.Block]types.Transactions) {
	for b, txs := range repMap {
		genTxs, ok := genMap[b]
		require.True(t, ok)
		require.Equal(t, len(txs), len(genTxs))
		for i, tx := range txs {
			require.Equal(t, tx.Hash().Hex(), genTxs[i].Hash().Hex())
		}
	}
}

type testParams struct {
	t            *testing.T
	initEvmBlock *evmcore.EvmBlock
	procEvmBlock *evmcore.EvmBlock
	initReceipts types.Receipts
	procReceipts types.Receipts
}

func newTestParams(t *testing.T, initEvmBlock, procEvmBlock *evmcore.EvmBlock, initReceipts, procReceipts types.Receipts) testParams {
	return testParams{t, initEvmBlock, procEvmBlock, initReceipts, procReceipts}
}

func (p testParams) compareEvmBlocks() {
	// comparing all fields of initEvmBlock and procEvmBlock
	require.Equal(p.t, p.initEvmBlock.Number, p.procEvmBlock.Number)
	//require.Equal(initEvmBlock.Hash, procEvmBlock.Hash)
	require.Equal(p.t, p.initEvmBlock.ParentHash, p.procEvmBlock.ParentHash)
	require.Equal(p.t, p.initEvmBlock.Root, p.procEvmBlock.Root)
	require.Equal(p.t, p.initEvmBlock.TxHash, p.procEvmBlock.TxHash)
	require.Equal(p.t, p.initEvmBlock.Time, p.procEvmBlock.Time)
	require.Equal(p.t, p.initEvmBlock.GasLimit, p.procEvmBlock.GasLimit)
	require.Equal(p.t, p.initEvmBlock.GasUsed, p.procEvmBlock.GasUsed)
	require.Equal(p.t, p.initEvmBlock.BaseFee, p.procEvmBlock.BaseFee)
}

func (p testParams) compareReceipts() {
	require.Equal(p.t, len(p.initReceipts), len(p.procReceipts))
	// compare every field except logs, I compare them separately
	for i, initRec := range p.initReceipts {
		require.Equal(p.t, initRec.BlockHash.String(), p.procReceipts[i].BlockHash.String())
		require.Equal(p.t, initRec.BlockNumber, p.procReceipts[i].BlockNumber)
		// TODO initRec.Bloom byte slices do not match
		// p.t.Log("initRec.Bloom.Bytes()", string(initRec.Bloom.Bytes())) ecxpected: empty string
		// p.t.Log("p.procReceipts[i].Bloom.Bytes()", string(p.procReceipts[i].Bloom.Bytes())) actual: @H
		//require.Equal(p.t, hexutils.BytesToHex(initRec.Bloom.Bytes()), hexutils.BytesToHex(p.procReceipts[i].Bloom.Bytes())) // TODO fix it do not match
		require.Equal(p.t, initRec.ContractAddress.Hex(), p.procReceipts[i].ContractAddress.Hex())
		require.Equal(p.t, initRec.CumulativeGasUsed, p.procReceipts[i].CumulativeGasUsed)
		require.Equal(p.t, hexutils.BytesToHex(initRec.PostState), hexutils.BytesToHex(p.procReceipts[i].PostState))
		require.Equal(p.t, initRec.Status, p.procReceipts[i].Status)
		require.Equal(p.t, initRec.TransactionIndex, p.procReceipts[i].TransactionIndex)
		require.Equal(p.t, initRec.TxHash.Hex(), p.procReceipts[i].TxHash.Hex())
		require.Equal(p.t, initRec.Type, p.procReceipts[i].Type)
	}
}

// TODO replace this func with checkLogsEquality from compareLogsByFilterCriteria
func (p testParams) compareLogs(initLogs2D, procLogs2D [][]*types.Log) {
	require.Equal(p.t, len(initLogs2D), len(procLogs2D))
	for i, initLogs := range initLogs2D {
		for j, initLog := range initLogs {
			// compare all fields
			require.Equal(p.t, initLog.Address.Hex(), procLogs2D[i][j].Address.Hex())
			require.Equal(p.t, initLog.BlockHash.Hex(), procLogs2D[i][j].BlockHash.Hex())
			require.Equal(p.t, initLog.BlockNumber, procLogs2D[i][j].BlockNumber)
			require.Equal(p.t, hexutils.BytesToHex(initLog.Data), hexutils.BytesToHex(procLogs2D[i][j].Data))
			require.Equal(p.t, initLog.Index, procLogs2D[i][j].Index)
			require.Equal(p.t, initLog.Removed, procLogs2D[i][j].Removed)

			for k, topic := range initLog.Topics {
				require.Equal(p.t, topic.Hex(), procLogs2D[i][j].Topics[k].Hex())
			}

			require.Equal(p.t, initLog.TxHash.Hex(), procLogs2D[i][j].TxHash.Hex())
			require.Equal(p.t, initLog.TxIndex, procLogs2D[i][j].TxIndex)
		}
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

// TODO consider to put initiator and processor on testParams
func (p testParams) compareTransactions(initiator, processor *testEnv) {
	ctx := context.Background()
	require.Equal(p.t, len(p.initEvmBlock.Transactions), len(p.procEvmBlock.Transactions))
	for i, tx := range p.initEvmBlock.Transactions {
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
	generator *testEnv
	processor *testEnv
	bvs []*inter.LlrSignedBlockVotes
	blockIndices []idx.Block
	epochToEvsMap map[idx.Epoch][]*inter.LlrSignedEpochVote
	t *testing.T
}

func newRepeater(s *IntegrationTestSuite) repeater {
	return repeater{
		generator: s.generator,
		processor: s.processor,
		bvs: s.bvs,
		blockIndices: s.blockIndices,
		epochToEvsMap: s.epochToEvsMap,
		t: s.T(),
	}
}

// processBlockVotesRecords processes block votes. Moreover, it processes block records for evert block index that has minimum 4 LLr Votes.
// Depending on 
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

// compareParams compares different parameters such as BlockByHash, BlockByNumber, Receipts, Logs
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

		// TODO handle this , testParams.serializeAndCompare(initReceipts, procReceipts) fails, receipts do not match
		// testParams.serializeAndCompare(initReceipts, procReceipts)
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
		// TODO compare logs fields
		testParams.serializeAndCompare(genLogs, procLogs) // test passes ok
		testParams.compareLogs(genLogs, procLogs)
		//testParams.compareLogsByQueries(ctx, initiator, processor)

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
	require.NoError(r.t,err)
	require.NotNil(r.t, lastLogs)

	defaultCrit := filters.FilterCriteria{FromBlock: big.NewInt(1), ToBlock: big.NewInt(int64(lastBlockNumber/2+1))}

	r.t.Log("compareLogsByFilterCriteria")
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

	r.t.Log("len(blockIndices)", len(r.blockIndices))




	checkLogsEquality := func(genLogs, procLogs []*types.Log) {
		// TODO s.Require().Equal(len(genLogs), len(procLogs))
		require.GreaterOrEqual(r.t, len(genLogs), len(procLogs))
		for i, procLog := range procLogs {
			// compare all fields
			require.Equal(r.t, procLog.Address.Hex(), genLogs[i].Address.Hex())
			require.Equal(r.t, procLog.BlockHash.Hex(), genLogs[i].BlockHash.Hex())
			require.Equal(r.t,procLog.BlockNumber, genLogs[i].BlockNumber)
			require.Equal(r.t,hexutils.BytesToHex(procLog.Data), hexutils.BytesToHex(genLogs[i].Data))
			require.Equal(r.t,procLog.Index, genLogs[i].Index)
			require.Equal(r.t,procLog.Removed, genLogs[i].Removed)

			for j, topic := range procLog.Topics {
				require.Equal(r.t, topic.Hex(), genLogs[i].Topics[j].Hex())
			}

			require.Equal(r.t,procLog.TxHash.Hex(), genLogs[i].TxHash.Hex())
			require.Equal(r.t,procLog.TxIndex, genLogs[i].TxIndex)
		}
	}

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
				require.NoError(r.t, err)
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

	r.t.Parallel()

	for _, tc := range testCases {
		tc := tc
		r.t.Run(tc.name, func(t *testing.T) {
			tc.pretest()
			genLogs, genErr := genApi.GetLogs(ctx, crit)
			procLogs, procErr := procApi.GetLogs(ctx, crit)
			if tc.success {
				require.NoError(t, procErr)
				require.NoError(t, genErr)
				checkLogsEquality(genLogs, procLogs)
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
				checkLogsEquality(genLogs, procLogs)
			}
		})
	}
}


func (s *IntegrationTestSuite) TestRepeater() {
	repeater := newRepeater(s)
	repeater.processEpochVotesRecords(s.startEpoch, s.lastEpoch)
	repeater.processBlockVotesRecords(true)

	s.Require().NoError(s.generator.store.Commit())
	s.Require().NoError(s.processor.store.Commit())

	// Compare transaction hashes
	s.T().Log("Checking procBlockToTxsMap <= genBlockToTxsMap")
	txByBlockSubsetOf(s.T(), s.procBlockToTxsMap, s.genBlockToTxsMap)

	// 2. Compare ER hashes
	repeater.compareERHashes(s.startEpoch+1, s.lastEpoch)

	s.T().Log("generator.BlockByNumber >= repeater.BlockByNumber")

	repeater.compareParams()

	repeater.compareLogsByFilterCriteria()
}

func (s *IntegrationTestSuite) TestFullRepeater() {

	repeater := newRepeater(s)

	wg := new(sync.WaitGroup)
	wg.Add(2)
	go func() {
		defer wg.Done()
		// process LLR epochVotes  in fullRepeater
		repeater.processEpochVotesRecords(s.startEpoch, s.lastEpoch)

		// process LLR block votes and BRs in fullReapeter
		repeater.processBlockVotesRecords(false)

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
					// TODO I might call processEvent here
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
	txByBlockSubsetOf(s.T(), s.genBlockToTxsMap, s.procBlockToTxsMap)

	// 2.Compare BlockByNumber,BlockByhash, GetReceipts, GetLogs
	repeater.compareParams()

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

	repeater.compareLogsByFilterCriteria()
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}