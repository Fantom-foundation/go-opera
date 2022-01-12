package gossip

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/inter/ibr"
	"github.com/Fantom-foundation/go-opera/inter/ier"
	"github.com/Fantom-foundation/go-opera/utils"

	"github.com/Fantom-foundation/lachesis-base/inter/idx"
)

// WIP test for ProcessFullBlockRecord and ProcessFullEpochRecord
func TestLLRCallbacks(t *testing.T) {

	/*
			Plan
			1. generate over 50 blocks using applyTx in generator
			2. retriving br and er GetFullBlockRecord and GetFullEpochRecord from generator
			3. Retrieve LLR votes  by iterating over tables s.table.LlrBlockVotes и s.table.LlrEpochVotes from generator
			4. Set llr votes on repeater how?
			5. conveting br and er in LlrIdxFullBlockRecord and LlrIdxEpochBlockRecord, epoch from one to last, block from one to last
		    6. run ProcessFullBlockRecord ProcessFullEpochRecord on repeater
		    7. compare parameters if generator and repeater using
			(b *EthAPIBackend) BlockByHash
		(b *EthAPIBackend) GetReceiptsByNumber
		(b *EthAPIBackend) GetReceipts
		(b *EthAPIBackend) GetLogs
		(b *EthAPIBackend) GetTransaction
		(f *Filter) Logs
		(s *Store) GetHistoryBlockEpochState

	*/

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

	// 2. retrieving br and er from generator done
	// 3 Retrieve LLR votes  by iterating over tables s.table.LlrBlockVotes и s.table.LlrEpochVotes from generator done

	bes := generator.store.getBlockEpochState()
	bvs := make([]*inter.LlrSignedBlockVotes, 0, bes.BlockState.LastBlock.Idx)
	iterateOverBlockVotes := func(env *testEnv, bvs []*inter.LlrSignedBlockVotes) {
		it := env.store.table.LlrBlockVotes.NewIterator(nil, nil)
		defer it.Release()
		for it.Next() {
			bv := &inter.LlrSignedBlockVotes{}
			if err := rlp.DecodeBytes(it.Value(), bv); err != nil {
				generator.store.Log.Crit("Failed to decode block vote", "err", err)
			}

			bvs = append(bvs, bv)
		}
	}
	iterateOverBlockVotes(generator, bvs)

	ibrs := make([]ibr.LlrIdxFullBlockRecord, 0, bes.BlockState.LastBlock.Idx)
	for b := idx.Block(1); b <= bes.BlockState.LastBlock.Idx; b++ {
		br := generator.store.GetFullBlockRecord(b)
		if br != nil {
			ibr := ibr.LlrIdxFullBlockRecord{LlrFullBlockRecord: *br, Idx: b}
			ibrs = append(ibrs, ibr)
		}
	}

	// TODO make sure bes.EpochState.Epoch is the last epoch
	evs := make([]*inter.LlrSignedEpochVote, 0, bes.EpochState.Epoch)
	// iterate over all records of LlrEpochVotes table and fetch all values
	iterateOverEpochVotes := func(env *testEnv, evs []*inter.LlrSignedEpochVote) {
		it := env.store.table.LlrEpochVotes.NewIterator(nil, nil)
		defer it.Release()
		for it.Next() {
			ev := &inter.LlrSignedEpochVote{}
			if err := rlp.DecodeBytes(it.Value(), ev); err != nil {
				generator.store.Log.Crit("Failed to decode epoch vote", "err", err)
			}

			evs = append(evs, ev)
		}
	}
	iterateOverEpochVotes(generator, evs)

	iers := make([]ier.LlrIdxFullEpochRecord, 0, bes.EpochState.Epoch)
	for e := idx.Epoch(startEpoch); e <= bes.EpochState.Epoch; e++ {
		er := generator.store.GetFullEpochRecord(e)
		if er != nil {
			ier := ier.LlrIdxFullEpochRecord{LlrFullEpochRecord: *er, Idx: e}
			iers = append(iers, ier)
		}
	}

	// 4. Set llr votes on repeater

	//creating repeater
	repeater := newTestEnv(startEpoch, validatorsNum)
	defer repeater.Close()

	for _, ev := range evs {
		require.NoError(repeater.ProcessEpochVote(*ev))
	}

	for _, bv := range bvs {
		require.NoError(repeater.ProcessBlockVotes(*bv))
	}

	// 6.run ProcessFullBlockRecord ProcessFullEpochRecord on repeater

	for _, ibr := range ibrs {
		require.NoError(repeater.ProcessFullBlockRecord(ibr))
	}

	for _, ier := range iers {
		require.NoError(repeater.ProcessFullEpochRecord(ier))
	}

	// 7. compare parameters of generator and repeater u
	/*
		(b *EthAPIBackend) BlockByHash
		(b *EthAPIBackend) GetReceiptsByNumber
		(b *EthAPIBackend) GetReceipts
		(b *EthAPIBackend) GetLogs
		(b *EthAPIBackend) GetTransaction
		(f *Filter) Logs
		(s *Store) GetHistoryBlockEpochState
	*/
	genBs, genEs := generator.store.GetHistoryBlockEpochState(startEpoch)
	repBs, repEs := repeater.store.GetHistoryBlockEpochState(startEpoch)
	require.Equal(genBs, repBs) // not sure
	require.Equal(genEs, repEs) // not sure
	// TODO do more comparisons

	// TODO add fullrepeater and make 2 go routines and then check if the states of 2 dbs match.
	//we make full repeater where we receviing events and llr votes from generator concurrently using two go routines
    // make sure that fullrepeater db state and geenrator db state is identical 
	
	/*
	fullReapeter := newTestEnv(startEpoch, validatorsNum)
	defer fullReapeter.Close()
	go func(){

	}()


	go  func(){

	}()
	*/


	/* testing ProcessEpochVote and ProcessBlockVote
	   1. make new processor testenv
	   2. loop over all evs and run processor.ProcessEpochVote(ev)
	   3. loop over all bvs and run processor.ProcessBlockVote(bv)
	   4. iterate over processor.store.table.LlrEpochVotes table and fetch all saved epoch votes
	   5. iterate over processor.store.table.LlrBlockVotes table and fetch all saved block votes
	   6. check whether fetched records match evs and bvs.
	*/

	//creating generator
	processor := newTestEnv(startEpoch, validatorsNum)
	defer processor.Close()

	for _, ev := range evs {
		require.NoError(processor.ProcessEpochVote(*ev))
	}

	for _, bv := range bvs {
		require.NoError(processor.ProcessBlockVotes(*bv))
	}

	blockVotes := make([]*inter.LlrSignedBlockVotes, 0, bes.BlockState.LastBlock.Idx)
	epochVotes := make([]*inter.LlrSignedEpochVote, 0, bes.EpochState.Epoch)
	iterateOverBlockVotes(processor, blockVotes)
	iterateOverEpochVotes(processor, epochVotes)

	require.Equal(len(blockVotes), len(bvs))
	require.Equal(len(epochVotes), len(evs))

	for i := 0; i < len(bvs); i++ {
		require.Equal(blockVotes[i].Signed.Size(), bvs[i].Signed.Size())
		require.Equal(blockVotes[i].Val.Hash().String(), bvs[i].Val.Hash().String()) // ?
		require.Equal(blockVotes[i].CalcPayloadHash().String(), bvs[i].CalcPayloadHash().String())
	}

	for i := 0; i < len(evs); i++ {
		require.Equal(epochVotes[i].Signed.Size(),  evs[i].Signed.Size())
		require.Equal(epochVotes[i].Val.Hash().String(), evs[i].Val.Hash().String()) // ?
	}











	

	

}

// TODO make sure there are no race conditions
// TODO plan
func TestProcessFullRecord2ThreadSafety(t *testing.T) {

}
