package gossip

import (
	"bytes"
	"context"
	"math/rand"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/Fantom-foundation/go-opera/evmcore"
	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/inter/ibr"
	"github.com/Fantom-foundation/go-opera/inter/ier"
	"github.com/Fantom-foundation/go-opera/utils"

	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/kvdb"

	"github.com/status-im/keycard-go/hexutils"
)

type IntegrationTestSuite struct {
	suite.Suite

	startEpoch idx.Epoch
	generator  *testEnv
	processor  *testEnv
}

// TODO add godoc
func (s *IntegrationTestSuite) SetupTest() {
	const (
		rounds        = 60
		validatorsNum = 10
		startEpoch    = 1
	)

	//creating generator and processor
	generator := newTestEnv(startEpoch, validatorsNum)
	processor := newTestEnv(startEpoch, validatorsNum)

	// generate txs and multiple blocks
	for n := uint64(0); n < rounds; n++ {
		// transfers
		txs := make([]*types.Transaction, validatorsNum)
		for i := idx.Validator(0); i < validatorsNum; i++ {
			from := i % validatorsNum
			to := 0
			txs[i] = generator.Transfer(idx.ValidatorID(from+1), idx.ValidatorID(to+1), utils.ToFtm(100))
		}
		tm := sameEpoch
		if n%10 == 0 {
			tm = nextEpoch
		}
		_, err := generator.ApplyTxs(tm, txs...)
		s.Require().NoError(err)
	}

	s.startEpoch = startEpoch
	s.generator = generator
	s.processor = processor

}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down test suite")
	s.generator.Close()
	s.processor.Close()
}

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

func processEpochVotesRecords(t *testing.T, epochToEvsMap map[idx.Epoch][]*inter.LlrSignedEpochVote, generator, processor *testEnv, startEpoch, lastEpoch idx.Epoch) {
	// invoke repeater.ProcessEpochVote and ProcessFullEpochRecord for epoch in range [2; lastepoch]
	for e := idx.Epoch(startEpoch + 1); e <= lastEpoch; e++ {
		epochVotes, ok := epochToEvsMap[e]
		if !ok {
			processor.store.Log.Crit("Failed to fetch epoch votes for a given epoch")
		}

		for _, v := range epochVotes {
			require.NoError(t, processor.ProcessEpochVote(*v))
		}

		if er := generator.store.GetFullEpochRecord(e); er != nil {
			ier := ier.LlrIdxFullEpochRecord{LlrFullEpochRecord: *er, Idx: e}
			require.NoError(t, processor.ProcessFullEpochRecord(ier))
		}
	}
}

func fetchBvsBlockIdxs(generator *testEnv) ([]*inter.LlrSignedBlockVotes, []idx.Block) {

	var (
		blockIdxCountMap map[idx.Block]uint64
		bvs              []*inter.LlrSignedBlockVotes
	)

	blockIdxCountMap = make(map[idx.Block]uint64)

	// fetching blockIdxs with at least minVoteCount
	fetchBlockIdxs := func(blockIdxCountMap map[idx.Block]uint64) (blockIdxs []idx.Block) {
		const minVoteCount = 4
		for blockIdx, count := range blockIdxCountMap {
			if count >= minVoteCount {
				blockIdxs = append(blockIdxs, blockIdx)
			}
		}
		return
	}

	// compute how any votes have been given for a particular block idx
	fillblockIdxCountMap := func(bv *inter.LlrSignedBlockVotes) {
		start, end := bv.Val.Start, bv.Val.Start+idx.Block(len(bv.Val.Votes))-1
		// check case if bv.Val.Votes == 0
		if start == end {
			blockIdxCountMap[start] += 1
			return
		}

		for start <= end {
			blockIdxCountMap[start] += 1
			start++
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

func processBlockVotesRecords(t *testing.T, bvs []*inter.LlrSignedBlockVotes, blockIdxs []idx.Block, generator, processor *testEnv) {
	for _, bv := range bvs {
		processor.ProcessBlockVotes(*bv)
	}

	for _, blockIdx := range blockIdxs {
		if br := generator.store.GetFullBlockRecord(blockIdx); br != nil {
			ibr := ibr.LlrIdxFullBlockRecord{LlrFullBlockRecord: *br, Idx: blockIdx}
			require.NoError(t, processor.ProcessFullBlockRecord(ibr))
		} else {
			generator.Log.Crit("Empty full block record popped up")
		}
	}
}

// 2a compare different parameters such as BlockByHash, BlockByNumber, Receipts, Logs
func compareParams(t *testing.T, blockIdxs []idx.Block, initiator, processor *testEnv) {

	ctx := context.Background()
	// compare blockbyNumber
	for _, blockIdx := range blockIdxs {

		// comparing EvmBlock by calling BlockByHash
		initEvmBlock, err := initiator.EthAPI.BlockByNumber(ctx, rpc.BlockNumber(blockIdx))
		require.NotNil(t, initEvmBlock)
		require.NoError(t, err)

		procEvmBlock, err := processor.EthAPI.BlockByNumber(ctx, rpc.BlockNumber(blockIdx))
		require.NotNil(t, procEvmBlock)
		require.NoError(t, err)

		// compare Receipts
		initReceipts := initiator.store.evm.GetReceipts(blockIdx, initiator.EthAPI.signer, initEvmBlock.Hash, initEvmBlock.Transactions)
		require.NotNil(t, initReceipts)
		procReceipts := processor.store.evm.GetReceipts(blockIdx, processor.EthAPI.signer, procEvmBlock.Hash, procEvmBlock.Transactions)
		require.NotNil(t, procReceipts)

		testParams := newTestParams(t, initEvmBlock, procEvmBlock, initReceipts, procReceipts)
		testParams.compareEvmBlocks()
		testParams.serializeAndCompare(initReceipts, procReceipts)

		// comparing evmBlock by calling BlockByHash
		// TODO should I compare of all Blocks or only block indexes for what 1/3W+1 votes have been given
		initEvmBlock, err = initiator.EthAPI.BlockByHash(ctx, initEvmBlock.Hash)
		require.NotNil(t, initEvmBlock)
		require.NoError(t, err)
		procEvmBlock, err = processor.EthAPI.BlockByHash(ctx, procEvmBlock.Hash)
		require.NotNil(t, procEvmBlock)
		require.NoError(t, err)

		testParams = newTestParams(t, initEvmBlock, procEvmBlock, initReceipts, procReceipts)
		testParams.compareEvmBlocks()

		// compare Logs
		initLogs, err := initiator.EthAPI.GetLogs(ctx, initEvmBlock.Hash)
		require.NotNil(t, initLogs)
		require.NoError(t, err)

		procLogs, err := processor.EthAPI.GetLogs(ctx, initEvmBlock.Hash)
		require.NotNil(t, procLogs)
		require.NoError(t, err)

		testParams.serializeAndCompare(initLogs, procLogs)

		// compare ReceiptForStorage
		initBR := initiator.store.GetFullBlockRecord(blockIdx)
		procBR := processor.store.GetFullBlockRecord(blockIdx)

		testParams.serializeAndCompare(initBR.Receipts, procBR.Receipts)

		// compare BR hashes
		require.Equal(t, initBR.Hash().Hex(), procBR.Hash().Hex())

		// compare transactions
		testParams.compareTransactions(initiator, processor)

	}
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

func (p testParams) serializeAndCompare(val1, val2 interface{}) {
	// serialize val1 and val2
	buf1, err := rlp.EncodeToBytes(val1)
	require.NotNil(p.t, buf1)
	require.NoError(p.t, err)
	buf2, err := rlp.EncodeToBytes(val2)
	require.NotNil(p.t, buf2)
	require.NoError(p.t, err)

	// compare serialized representation of val1 and val2
	require.True(p.t, bytes.Equal(buf1, buf2))
}

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

func (s *IntegrationTestSuite) TestRepeater() {
	epochToEvsMap := fetchEvs(s.generator)
	lastEpoch := s.generator.store.GetEpoch()
	processEpochVotesRecords(s.T(), epochToEvsMap, s.generator, s.processor, s.startEpoch, lastEpoch)

	bvs, blockIdxs := fetchBvsBlockIdxs(s.generator)
	processBlockVotesRecords(s.T(), bvs, blockIdxs, s.generator, s.processor)

	s.Require().NoError(s.generator.store.Commit())
	s.Require().NoError(s.processor.store.Commit())

	genBlockToTxsMap := fetchTxsbyBlock(s.generator)
	repBlockToTxsMap := fetchTxsbyBlock(s.processor)

	// 1.Compare transaction hashes
	s.T().Log("Checking repBlockToTxsMap <= genBlockToTxsMap")
	txByBlockSubsetOf(s.T(), repBlockToTxsMap, genBlockToTxsMap)

	// 2. Compare ER hashes
	compareERHashes := func(startEpoch, lastEpoch idx.Epoch) {
		for e := startEpoch; e <= lastEpoch; e++ {

			genBs, genEs := s.generator.store.GetHistoryBlockEpochState(e)
			repBs, repEs := s.processor.store.GetHistoryBlockEpochState(e)
			s.Require().Equal(genBs.Hash().Hex(), repBs.Hash().Hex())
			s.Require().Equal(genEs.Hash().Hex(), repEs.Hash().Hex())

			genEr := s.generator.store.GetFullEpochRecord(e)
			repEr := s.processor.store.GetFullEpochRecord(e)
			s.Require().Equal(genEr.Hash().Hex(), repEr.Hash().Hex())
		}
	}

	compareERHashes(s.startEpoch+1, lastEpoch)

	s.T().Log("generator.BlockByNumber >= repeater.BlockByNumber")
	compareParams(s.T(), blockIdxs, s.generator, s.processor)
}

func (s *IntegrationTestSuite) TestFullRepeater() {

	bvs, blockIdxs := fetchBvsBlockIdxs(s.generator)
	epochToEvsMap := fetchEvs(s.generator)
	lastEpoch := s.generator.store.GetEpoch()

	wg := new(sync.WaitGroup)
	wg.Add(2)
	go func(epochToEvsMap map[idx.Epoch][]*inter.LlrSignedEpochVote, bvs []*inter.LlrSignedBlockVotes, blockIdxs []idx.Block) {
		defer wg.Done()
		// process LLR epochVotes  in fullRepeater
		processEpochVotesRecords(s.T(), epochToEvsMap, s.generator, s.processor, s.startEpoch, lastEpoch)

		// process LLR block votes and BRs in fullReapeter
		processBlockVotesRecords(s.T(), bvs, blockIdxs, s.generator, s.processor)

	}(epochToEvsMap, bvs, blockIdxs)

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
	genBlockToTxsMap := fetchTxsbyBlock(s.generator)
	fullRepBlockToTxsMap := fetchTxsbyBlock(s.processor)

	s.T().Log("Checking genBlockToTxsMap <= fullRepBlockToTxsMap")
	txByBlockSubsetOf(s.T(), genBlockToTxsMap, fullRepBlockToTxsMap)

	// 2.Compare BlockByNumber,BlockByhash, GetReceipts, GetLogs
	compareParams(s.T(), blockIdxs, s.generator, s.processor)

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
			s.Require().Equal(hexutils.BytesToHex(v), hexutils.BytesToHex(k))
		}
	}

	s.T().Log("Checking genKVs <= fullKVs")
	subsetOf(genKVMap, fullRepKVMap)

	// IndexLogs scenario
	// run EvmLogs.Push(l) for all logs and then compare the states of generator and processor
	ctx := context.Background()
	s.Require().NotNil(blockIdxs)

	evmBlock, err := s.processor.EthAPI.BlockByNumber(ctx, rpc.BlockNumber(blockIdxs[0]))
	s.Require().NotNil(evmBlock)
	s.Require().NoError(err)

	logs2D, err := s.processor.EthAPI.GetLogs(ctx, evmBlock.Hash)
	s.Require().NotNil(evmBlock)
	s.Require().NoError(err)

	testIndexLogs := func(logs2D [][]*types.Log) {
		for _, logs := range logs2D {
			for _, l := range logs {
				s.Require().NoError(s.processor.store.EvmStore().EvmLogs.Push(l))
			}
		}
	}

	testIndexLogs(logs2D)
	genKVMapAfterIndexLogs := fetchTable(s.generator.store.mainDB)
	fullRepKVMapAfterIndexLogs := fetchTable(s.processor.store.mainDB)

	// comparing the states
	subsetOf(genKVMap, genKVMapAfterIndexLogs)
	subsetOf(fullRepKVMap, fullRepKVMapAfterIndexLogs)
	subsetOf(genKVMapAfterIndexLogs, fullRepKVMapAfterIndexLogs)

	// Search logs by topic, block and address
	// make sure it work as expected
	// see  TestIndexSearchMultyVariants in topicsdb/topicsdb_test.go

	testSearchLogsWithLLRSync := func(blockIdx idx.Block) {

		randAddress := func() (addr common.Address) {
			n, err := rand.Read(addr[:])
			if err != nil {
				panic(err)
			}
			if n != common.AddressLength {
				panic("address is not filled")
			}
			return
		}

		var (
			hash1 = common.BytesToHash([]byte("topic1"))
			hash2 = common.BytesToHash([]byte("topic2"))
			hash3 = common.BytesToHash([]byte("topic3"))
			hash4 = common.BytesToHash([]byte("topic4"))
			addr1 = randAddress()
			addr2 = randAddress()
			addr3 = randAddress()
			addr4 = randAddress()
		)

		// I looked atso at Logs2D, if loop over them, Log struct has empty fields.
		// To resolve this issue, I came up with testData
		testdata := []*types.Log{{
			BlockNumber: uint64(blockIdx),
			Address:     addr1,
			Topics:      []common.Hash{hash1, hash1, hash1},
		}, {
			BlockNumber: uint64(blockIdx),
			Address:     addr2,
			Topics:      []common.Hash{hash2, hash2, hash2},
		}, {
			BlockNumber: uint64(blockIdx),
			Address:     addr3,
			Topics:      []common.Hash{hash3, hash3, hash3},
		}, {
			BlockNumber: uint64(blockIdx),
			Address:     addr4,
			Topics:      []common.Hash{hash4, hash4, hash4},
		},
		}

		index := s.processor.store.EvmStore().EvmLogs

		for _, l := range testdata {
			s.Require().NoError(index.Push(l))
		}

		// require.ElementsMatchf(testdata, got, "") doesn't work properly here,
		// so use check()
		check := func(got []*types.Log) {
			// why we declared count here?
			count := 0
			for _, a := range got {
				for _, b := range testdata {
					if b.Address == a.Address {
						s.Require().ElementsMatch(a.Topics, b.Topics)
						count++
						break
					}
				}
			}
		}

		for dsc, method := range map[string]func(context.Context, idx.Block, idx.Block, [][]common.Hash) ([]*types.Log, error){
			"sync": index.FindInBlocks,
			//	"async": index.FindInBlocksAsync,
		} {
			s.Run(dsc, func() {

				s.Run("With no addresses", func() {
					got, err := method(nil, 0, 1000, [][]common.Hash{
						{},
						{hash1, hash2, hash3, hash4},
						{},
						{hash1, hash2, hash3, hash4},
					})
					s.Require().NoError(err)
					//s.Require().Equal(4, len(got))
					check(got)
				})

				s.Run("With addresses", func() {
					got, err := method(nil, 0, 1000, [][]common.Hash{
						{addr1.Hash(), addr2.Hash(), addr3.Hash(), addr4.Hash()},
						{hash1, hash2, hash3, hash4},
						{},
						{hash1, hash2, hash3, hash4},
					})
					s.Require().NoError(err)
					//s.Require().Equal(4, len(got))
					check(got)
				})

				s.Run("With block range", func() {
					got, err := method(nil, 2, 998, [][]common.Hash{
						{addr1.Hash(), addr2.Hash(), addr3.Hash(), addr4.Hash()},
						{hash1, hash2, hash3, hash4},
						{},
						{hash1, hash2, hash3, hash4},
					})
					s.Require().NoError(err)
					//s.Require().Equal(2, len(got))
					check(got)
				})

				s.Run("With addresses and blocks", func() {
					got1, err := method(nil, 2, 998, [][]common.Hash{
						{addr1.Hash(), addr2.Hash(), addr3.Hash(), addr4.Hash()},
						{hash1, hash2, hash3, hash4},
						{},
						{hash1, hash2, hash3, hash4},
					})
					s.Require().NoError(err)
					//s.Require().Equal(2, len(got1))
					check(got1)

					got2, err := method(nil, 2, 998, [][]common.Hash{
						{addr4.Hash(), addr3.Hash(), addr2.Hash(), addr1.Hash()},
						{hash1, hash2, hash3, hash4},
						{},
						{hash1, hash2, hash3, hash4},
					})
					s.Require().NoError(err)
					s.Require().ElementsMatch(got1, got2)
				})

			})

		}
	}

	testSearchLogsWithLLRSync(blockIdxs[0])
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

/*

func TestLLRCallbacks(t *testing.T) {
	const (
		rounds        = 60
		validatorsNum = 10
		startEpoch    = 1
	)

	require := require.New(t)

	//creating generator
	generator := newTestEnv(startEpoch, validatorsNum)
	defer generator.Close()

	// generate txs and multiple blocks
	for n := uint64(0); n < rounds; n++ {
		// transfers
		txs := make([]*types.Transaction, validatorsNum)
		for i := idx.Validator(0); i < validatorsNum; i++ {
			from := i % validatorsNum
			to := 0
			txs[i] = generator.Transfer(idx.ValidatorID(from+1), idx.ValidatorID(to+1), utils.ToFtm(100))
		}
		tm := sameEpoch
		if n%10 == 0 {
			tm = nextEpoch
		}
		_, err := generator.ApplyTxs(tm, txs...)
		require.NoError(err)
	}

	fetchEvs := func(env *testEnv) map[idx.Epoch][]*inter.LlrSignedEpochVote {
		m := make(map[idx.Epoch][]*inter.LlrSignedEpochVote)
		it := env.store.table.LlrEpochVotes.NewIterator(nil, nil)
		defer it.Release()
		for it.Next() {
			ev := &inter.LlrSignedEpochVote{}
			if err := rlp.DecodeBytes(it.Value(), ev); err != nil {
				env.store.Log.Crit("Failed to decode epoch vote", "err", err)
			}

			if ev != nil {
				m[ev.Val.Epoch] = append(m[ev.Val.Epoch], ev)
			}
		}
		return m
	}

	epochToEvsMap := fetchEvs(generator)
	lastEpoch := generator.store.GetEpoch()
	t.Log("lastEpoch", lastEpoch)
	lastBlock := generator.store.GetBlockState().LastBlock.Idx
	t.Log("lastBlock", lastBlock)

	// create repeater
	repeater := newTestEnv(startEpoch, validatorsNum)
	defer repeater.Close()

	processEpochVotesRecords := func(epochToEvsMap map[idx.Epoch][]*inter.LlrSignedEpochVote, processor *testEnv) {
		// invoke repeater.ProcessEpochVote and ProcessFullEpochRecord for epoch in range [2; lastepoch]
		for e := idx.Epoch(startEpoch + 1); e <= lastEpoch; e++ {
			epochVotes, ok := epochToEvsMap[e]
			if !ok {
				repeater.store.Log.Crit("Failed to fetch epoch votes for a given epoch")
			}

			for _, v := range epochVotes {
				require.NoError(processor.ProcessEpochVote(*v))
			}

			if er := generator.store.GetFullEpochRecord(e); er != nil {
				ier := ier.LlrIdxFullEpochRecord{LlrFullEpochRecord: *er, Idx: e}
				require.NoError(processor.ProcessFullEpochRecord(ier))
			}
		}
	}

	processEpochVotesRecords(epochToEvsMap, repeater)

	fetchBvsBlockIdxs := func() ([]*inter.LlrSignedBlockVotes, []idx.Block) {

		var (
			blockIdxCountMap map[idx.Block]uint64
			bvs              []*inter.LlrSignedBlockVotes
		)

		blockIdxCountMap = make(map[idx.Block]uint64)

		// fetching blockIdxs with at least minVoteCount
		fetchBlockIdxs := func(blockIdxCountMap map[idx.Block]uint64) (blockIdxs []idx.Block) {
			const minVoteCount = 4
			for blockIdx, count := range blockIdxCountMap {
				if count >= minVoteCount {
					blockIdxs = append(blockIdxs, blockIdx)
				}
			}
			return
		}

		// compute how any votes have been given for a particular block idx
		fillblockIdxCountMap := func(bv *inter.LlrSignedBlockVotes) {
			start, end := bv.Val.Start, bv.Val.Start+idx.Block(len(bv.Val.Votes))-1
			// check case if bv.Val.Votes == 0
			if start == end {
				blockIdxCountMap[start] += 1
				return
			}

			for start <= end {
				blockIdxCountMap[start] += 1
				start++
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

	// fetch LLRBlockVotes and blockIdxs with at least 4 Votes
	bvs, blockIdxs := fetchBvsBlockIdxs()

	processBlockVotesRecords := func(bvs []*inter.LlrSignedBlockVotes, blockIdxs []idx.Block, processor *testEnv) {
		for _, bv := range bvs {
			processor.ProcessBlockVotes(*bv)
		}

		for _, blockIdx := range blockIdxs {
			if br := generator.store.GetFullBlockRecord(blockIdx); br != nil {
				ibr := ibr.LlrIdxFullBlockRecord{LlrFullBlockRecord: *br, Idx: blockIdx}
				require.NoError(processor.ProcessFullBlockRecord(ibr))
			} else {
				generator.Log.Crit("Empty full block record popped up")
			}
		}
	}

	// process all LLR Block Votes and BRs for blockIdxs with at least 4 Votes
	processBlockVotesRecords(bvs, blockIdxs, repeater)

	require.NoError(generator.store.Commit())
	require.NoError(repeater.store.Commit())

	// Compare the states of generator and repeater

	fetchTxsbyBlock := func(env *testEnv) map[idx.Block]types.Transactions {
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

	genBlockToTxsMap := fetchTxsbyBlock(generator)
	repBlockToTxsMap := fetchTxsbyBlock(repeater)

	txByBlockSubsetOf := func(repMap, genMap map[idx.Block]types.Transactions) {
		for b, txs := range repMap {
			genTxs, ok := genMap[b]
			require.True(ok)
			require.Equal(len(txs), len(genTxs))
			for i, tx := range txs {
				require.Equal(tx.Hash().Hex(), genTxs[i].Hash().Hex())
			}
		}
	}

	// 1.Compare transaction hashes
	t.Log("Checking repBlockToTxsMap <= genBlockToTxsMap")
	txByBlockSubsetOf(repBlockToTxsMap, genBlockToTxsMap)

	// 2.BlockByNUmber
	// func (b *EthAPIBackend) BlockByNumber(ctx context.Context, number rpc.BlockNumber) (*evmcore.EvmBlock, error) {
	//

	// 2.Compare  ER hashes
	for e := idx.Epoch(2); e <= lastEpoch; e++ {

		genBs, genEs := generator.store.GetHistoryBlockEpochState(e)
		repBs, repEs := repeater.store.GetHistoryBlockEpochState(e)
		require.Equal(genBs.Hash().Hex(), repBs.Hash().Hex())
		require.Equal(genEs.Hash().Hex(), repEs.Hash().Hex())

		genEr := generator.store.GetFullEpochRecord(e)
		repEr := repeater.store.GetFullEpochRecord(e)
		require.Equal(genEr.Hash().Hex(), repEr.Hash().Hex())
	}

	// 2a compare different parameters such as BlockByHash, BlockByNumber, Receipts, Logs
	compareParams := func(blockIdxs []idx.Block, initiator, processor *testEnv) {
		// initiator is generator
		// processor is ether fullRep or repeater
		ctx := context.Background()

		compareEvmBlocks := func(initEvmBlock, procEvmBlock *evmcore.EvmBlock) {
			// comparing all fields of initEvmBlock and procEvmBlock
			require.Equal(initEvmBlock.Number, procEvmBlock.Number)
			//require.Equal(initEvmBlock.Hash, procEvmBlock.Hash)
			require.Equal(initEvmBlock.ParentHash, procEvmBlock.ParentHash)
			require.Equal(initEvmBlock.Root, procEvmBlock.Root)
			require.Equal(initEvmBlock.TxHash, procEvmBlock.TxHash)
			require.Equal(initEvmBlock.Time, procEvmBlock.Time)
			require.Equal(initEvmBlock.GasLimit, procEvmBlock.GasLimit)
			require.Equal(initEvmBlock.GasUsed, procEvmBlock.GasUsed)
			require.Equal(initEvmBlock.BaseFee, procEvmBlock.BaseFee)
			require.Equal(len(initEvmBlock.Transactions), len(procEvmBlock.Transactions))
			for i, tx := range initEvmBlock.Transactions {
				require.Equal(tx.Hash().Hex(), procEvmBlock.Transactions[i].Hash().Hex())
			}
		}

		serializeAndCompare := func(val1, val2 interface{}) {
			// serialize val1 and val2
			buf1, err := rlp.EncodeToBytes(val1)
			require.NotNil(buf1)
			require.NoError(err)
			buf2, err := rlp.EncodeToBytes(val2)
			require.NotNil(buf2)
			require.NoError(err)

			// compare serialized representation of val1 and val2
			require.True(bytes.Equal(buf1, buf2))
		}

		// compare blockbyNumber
		for _, blockIdx := range blockIdxs {

			// comparing EvmBlock by calling BlockByHash
			initEvmBlock, err := initiator.EthAPI.BlockByNumber(ctx, rpc.BlockNumber(blockIdx))
			require.NotNil(initEvmBlock)
			require.NoError(err)

			procEvmBlock, err := processor.EthAPI.BlockByNumber(ctx, rpc.BlockNumber(blockIdx))
			require.NotNil(procEvmBlock)
			require.NoError(err)

			compareEvmBlocks(initEvmBlock, procEvmBlock)

			// comparing evmBlock by calling BlockByHash
			// TODO should I compare of all Blocks or only block indexes for what 1/3W+1 votes have been given
			initEvmBlock, err = initiator.EthAPI.BlockByHash(ctx, initEvmBlock.Hash)
			require.NotNil(initEvmBlock)
			require.NoError(err)
			procEvmBlock, err = processor.EthAPI.BlockByHash(ctx, procEvmBlock.Hash)
			require.NotNil(procEvmBlock)
			require.NoError(err)

			compareEvmBlocks(initEvmBlock, procEvmBlock)

			// compare Receipts
			initReceipts := initiator.store.evm.GetReceipts(blockIdx, initiator.EthAPI.signer, initEvmBlock.Hash, initEvmBlock.Transactions)
			require.NotNil(initReceipts)
			procReceipts := processor.store.evm.GetReceipts(blockIdx, processor.EthAPI.signer, procEvmBlock.Hash, procEvmBlock.Transactions)
			require.NotNil(procReceipts)

			serializeAndCompare(initReceipts, procReceipts)

			// TODO compare indexLogs

			// compare Logs
			initLogs, err := initiator.EthAPI.GetLogs(ctx, initEvmBlock.Hash)
			require.NotNil(initLogs)
			require.NoError(err)

			procLogs, err := processor.EthAPI.GetLogs(ctx, initEvmBlock.Hash)
			require.NotNil(procLogs)
			require.NoError(err)

			serializeAndCompare(initLogs, procLogs)

			// compare ReceiptForStorage
			initBR := initiator.store.GetFullBlockRecord(blockIdx)
			procBR := processor.store.GetFullBlockRecord(blockIdx)

			serializeAndCompare(initBR.Receipts, procBR.Receipts)

			// compare BR hashes
			require.Equal(initBR.Hash().Hex(), procBR.Hash().Hex())
		}
	}

	t.Log("generator.BlockByNumber >= repeater.BlockByNumber")
	compareParams(blockIdxs, generator, repeater)

	// declare fullRepeater
	fullRepeater := newTestEnv(startEpoch, validatorsNum)
	defer fullRepeater.Close()

	wg := new(sync.WaitGroup)
	wg.Add(2)
	go func(fullRepeater *testEnv, epochToEvsMap map[idx.Epoch][]*inter.LlrSignedEpochVote, bvs []*inter.LlrSignedBlockVotes, blockIdxs []idx.Block) {
		defer wg.Done()
		// process LLR epochVotes  in fullRepeater
		processEpochVotesRecords(epochToEvsMap, fullRepeater)

		// process LLR block votes and BRs in fullReapeter
		processBlockVotesRecords(bvs, blockIdxs, fullRepeater)

	}(fullRepeater, epochToEvsMap, bvs, blockIdxs)

	go func(fullRepeater *testEnv) {
		defer wg.Done()
		events := func() (events []*inter.EventPayload) {
			it := generator.store.table.Events.NewIterator(nil, nil)
			defer it.Release()
			for it.Next() {
				e := &inter.EventPayload{}
				if err := rlp.DecodeBytes(it.Value(), e); err != nil {
					generator.store.Log.Crit("Failed to decode event", "err", err)
				}
				if e != nil {
					// TODO I might call processEvent here
					events = append(events, e)
				}
			}
			return
		}()

		for _, e := range events {
			fullRepeater.engineMu.Lock()
			require.NoError(fullRepeater.processEvent(e))
			fullRepeater.engineMu.Unlock()
		}
	}(fullRepeater)

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

	require.NoError(fullRepeater.store.Commit())

	// Comparing generator and fullRepeater states

	// 1.Comparing Tx hashes
	fullRepBlockToTxsMap := fetchTxsbyBlock(fullRepeater)
	t.Log("Checking genBlockToTxsMap <= fullRepBlockToTxsMap")
	txByBlockSubsetOf(genBlockToTxsMap, fullRepBlockToTxsMap)

	// 2.Compare BlockByNumber
	compareParams(blockIdxs, generator, fullRepeater)

	// 2. Comparing mainDb of generator and fullRepeater
	genKVMap := fetchTable(generator.store.mainDB)
	fullRepKVMap := fetchTable(fullRepeater.store.mainDB)

	subsetOf := func(aa, bb map[string]string) {
		for _k, _v := range aa {
			k, v := []byte(_k), []byte(_v)
			if k[0] == 0 || k[0] == 'x' || k[0] == 'X' || k[0] == 'b' || k[0] == 'S' {
				continue
			}
			require.Equal(hexutils.BytesToHex(v), hexutils.BytesToHex([]byte(bb[_k])), hexutils.BytesToHex(k))
		}
	}

	t.Log("Checking genKVs <= fullKVs")
	subsetOf(genKVMap, fullRepKVMap)
}
*/
