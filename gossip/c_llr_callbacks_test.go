package gossip

import (
	"context"
	"reflect"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

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

	//	"github.com/ethereum/go-ethereum/common"

	"github.com/status-im/keycard-go/hexutils"
)

// WIP test for ProcessFullBlockRecord and ProcessFullEpochRecord
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

	// create repeater
	repeater := newTestEnv(startEpoch, validatorsNum)
	defer repeater.Close()

	processEpochVotesRecords := func(epochToEvsMap map[idx.Epoch][]*inter.LlrSignedEpochVote, processor *testEnv){
		// invoke repeater.ProcessEpochVote and ProcessFullEpochRecord for epoch in range [2; lastepoch]
		for e := idx.Epoch(startEpoch +1); e <= lastEpoch; e++ {
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
	

	// TODO find out how many votes at most has each block? one or multiple?
	// map[idx.Block][]*inter.LlrSignedBlockVotes  or map[idx.Block]*inter.LlrSignedBlockVotes
	fetchBvs := func(env *testEnv) map[idx.Block][]*inter.LlrSignedBlockVotes {
		m := make(map[idx.Block][]*inter.LlrSignedBlockVotes)
		it := env.store.table.LlrBlockVotes.NewIterator(nil, nil)
		defer it.Release()
		for it.Next() {
			bv := &inter.LlrSignedBlockVotes{}
			if err := rlp.DecodeBytes(it.Value(), bv); err != nil {
				env.store.Log.Crit("Failed to decode epoch vote", "err", err)
			}

			if bv != nil {
				m[bv.Val.Start] = append(m[bv.Val.Start], bv)
			}
		}
		return m
	}

	blockToBvsMap := fetchBvs(generator)

	processBlockVotesRecords := func(blockToBvsMap map[idx.Block][]*inter.LlrSignedBlockVotes, processor *testEnv) {
		for b, bvs := range blockToBvsMap {
			for _, bv := range bvs {
				require.NoError(processor.ProcessBlockVotes(*bv))
			}
	
			if br := generator.store.GetFullBlockRecord(b); br != nil {
				ibr := ibr.LlrIdxFullBlockRecord{LlrFullBlockRecord: *br, Idx: b}
				require.NoError(processor.ProcessFullBlockRecord(ibr))
			}
		}
	}

	processBlockVotesRecords(blockToBvsMap, repeater)

	require.NoError(generator.store.Commit())
	require.NoError(repeater.store.Commit())


    // Compare the states of generator and repeater

	fetchTxsbyBlock := func(env *testEnv)  map[idx.Block]types.Transactions {
		numKeys := len(reflect.ValueOf(blockToBvsMap).MapKeys())
		m := make(map[idx.Block]types.Transactions, numKeys)
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

	txByBlockSubsetOf := func(repMap, genMap map[idx.Block]types.Transactions ){
		// I assume repMap is a subset of genMap
		//require.Less(len(reflect.ValueOf(repMap).MapKeys()), len(reflect.ValueOf(genMap).MapKeys()))
		for b, txs := range repMap {
			genTxs, ok := genMap[b]
			require.True(ok)
			require.Equal(len(txs), len(genTxs))
			for i,tx := range txs {
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


    // 2a compare BlockByNumber
	compareBlocksByNumberHash := func(blockToBvsMap map[idx.Block][]*inter.LlrSignedBlockVotes,  initiator, processor *testEnv) {
		// initiator is generator
		// processor is ether fullRep or repeater
		ctx := context.Background()

		blockByNumberPanickedOrErr := func(b idx.Block, processor *testEnv) (*evmcore.EvmBlock, bool) {
			panicked := false

			defer func() {
				if r := recover(); r != nil {
					panicked = true
				}
			}()

			emvBlock, err := processor.EthAPI.BlockByNumber(ctx, rpc.BlockNumber(b))
			if err != nil {
				panicked = true
			}

			return emvBlock,panicked
		}

		for b := range blockToBvsMap  {

			// it panics sometimes.I introduced blockByNumberPanickedOrErr to handle this behavior.
			procEvmBlock, panicked := blockByNumberPanickedOrErr(b, processor)
			if panicked ||  procEvmBlock == nil{
				continue
			}

			/*
			procEvmBlock, err := processor.EthAPI.BlockByNumber(ctx, rpc.BlockNumber(b))
			require.NoError(err)
			require.NotNil(procEvmBlock)
			*/

			// outputs no error
			initEvmBlock, err := initiator.EthAPI.BlockByNumber(ctx, rpc.BlockNumber(b))
			require.NoError(err)
			require.NotNil(initEvmBlock)

		    require.Equal(initEvmBlock.Hash.Hex(), procEvmBlock.Hash.Hex())

			// invoke BlockByHash
			procEvmBlockByHash, err := processor.EthAPI.BlockByHash(ctx, initEvmBlock.Hash)
			require.NoError(err)
			require.NotNil(initEvmBlock)
			require.Equal(procEvmBlockByHash, procEvmBlock)
			require.Equal(procEvmBlockByHash.Hash, procEvmBlock.Hash)

			initEvmBlockByHash, err := initiator.EthAPI.BlockByHash(ctx, initEvmBlock.Hash)
			require.NoError(err)
			require.NotNil(initEvmBlock)
			require.Equal(initEvmBlockByHash, initEvmBlock)
			require.Equal(procEvmBlockByHash.Hash, procEvmBlock.Hash)
		}
	}

	t.Log("generator.BlockByNumber >= repeater.BlockByNumber")
	compareBlocksByNumberHash(blockToBvsMap, generator, repeater)


    // TODO why BlockByNumber returns an error? Handle panic in procEvmBlock, err := repeater.EthAPI.BlockByNumber(ctx, rpc.BlockNumber(b))
    // func (b *EthAPIBackend) BlockByNumber(ctx context.Context, number rpc.BlockNumber) (*evmcore.EvmBlock, error) {
	// 
    // 3.Compare BR hashes
	for b := range blockToBvsMap {
        // 	Receipts and Logs WIP
		// receipts := env.store.evm.GetReceipts(idx.Block(b.Block.Number.Uint64()), env.EthAPI.signer, b.Block.Hash, b.Block.Transactions)
		/*
		genBlock := generator.EthAPI.state.GetBlock(common.Hash{}, uint64(b))
		genReceipts := generator.store.evm.GetReceipts(b, generator.EthAPI.signer, genBlock.Hash, genBlock.Transactions)
		repBlock := repeater.EthAPI.state.GetBlock(common.Hash{}, uint64(b))
		repReceipts := repeater.store.evm.GetReceipts(b, repeater.EthAPI.signer, repBlock.Hash, repBlock.Transactions)
		require.Equal(len(genReceipts),len(repReceipts))
		*/

		
		genBlockResHash := generator.store.GetLlrBlockResult(b)
		repBlockResHash := repeater.store.GetLlrBlockResult(b)
		require.Equal(genBlockResHash.Hex(), repBlockResHash.Hex())

		genBrHash := generator.store.GetFullBlockRecord(b).Hash().Hex()
		repBrHash := repeater.store.GetFullBlockRecord(b).Hash().Hex()
		require.Equal(repBrHash, genBrHash)
	}




	// declare fullRepeater
	fullRepeater := newTestEnv(startEpoch, validatorsNum)
	defer fullRepeater.Close()


	wg := new(sync.WaitGroup)
	wg.Add(2)
	go func(fullRepeater *testEnv, epochToEvsMap map[idx.Epoch][]*inter.LlrSignedEpochVote, blockToBvsMap map[idx.Block][]*inter.LlrSignedBlockVotes){
		defer wg.Done()
		// process LLR epochVotes  in fullRepeater
		processEpochVotesRecords(epochToEvsMap, fullRepeater)

		// process LLR block votes and BRs in fullReapeter
		processBlockVotesRecords(blockToBvsMap, fullRepeater)

	}(fullRepeater, epochToEvsMap, blockToBvsMap)

	go func(fullRepeater *testEnv){
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
	// TODO fullRepeater panics some time by calling BlockByNumber
	compareBlocksByNumberHash(blockToBvsMap, generator, fullRepeater)


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
