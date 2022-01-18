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
	"github.com/Fantom-foundation/lachesis-base/kvdb"

	"github.com/status-im/keycard-go/hexutils"
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
	

	// invoke repeater.ProcessEpochVote and ProcessFullEpochRecord for epoch in range [2; lastepoch]
	for e := idx.Epoch(2); e <= lastEpoch; e++ {
		epochVotes, ok := epochToEvsMap[e]
		if !ok {
			repeater.store.Log.Crit("Failed to fetch epoch votes for a given epoch")
		}

		for _, v := range epochVotes {
			require.NoError(repeater.ProcessEpochVote(*v))
		}

		if er := generator.store.GetFullEpochRecord(e); er != nil {
			ier := ier.LlrIdxFullEpochRecord{LlrFullEpochRecord: *er, Idx: e}
			require.NoError(repeater.ProcessFullEpochRecord(ier))
		}
	}

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

	for b, bvs := range blockToBvsMap {
		for _, bv := range bvs {
			require.NoError(repeater.ProcessBlockVotes(*bv))
		}

		if br := generator.store.GetFullBlockRecord(b); br != nil {
			ibr := ibr.LlrIdxFullBlockRecord{LlrFullBlockRecord: *br, Idx: b}
			require.NoError(repeater.ProcessFullBlockRecord(ibr))
		}
	}

	// compare results
	// TODO check more parameters
	for e := idx.Epoch(2); e <= lastEpoch; e++ {
		genBs, genEs := generator.store.GetHistoryBlockEpochState(e)
		repBs, repEs := repeater.store.GetHistoryBlockEpochState(e)
		require.Equal(genBs.Hash().Hex(), repBs.Hash().Hex())
		require.Equal(genEs.Hash().Hex(), repEs.Hash().Hex())

		genEr := generator.store.GetFullEpochRecord(e)
		repEr := repeater.store.GetFullEpochRecord(e)
		require.Equal(genEr.Hash().Hex(), repEr.Hash().Hex())
	}

	for b := range blockToBvsMap {
		genBrHash := generator.store.GetFullBlockRecord(b).Hash().Hex()
		repBrHash := repeater.store.GetFullBlockRecord(b).Hash().Hex()
		require.Equal(repBrHash, genBrHash)
	}

	fullRepeater := newTestEnv(startEpoch, validatorsNum)
	defer fullRepeater.Close()


	// process LLR epochVotes  in fullRepeater
	// with epochVotes there is no need to run them concurrently, cause you encounter an error
	// TODO make a standalone func for that 
	for e := idx.Epoch(2); e <= lastEpoch; e++ {
		epochVotes, ok := epochToEvsMap[e]
		if !ok {
			repeater.store.Log.Crit("Failed to fetch epoch votes for a given epoch")
		}

		for _, v := range epochVotes {
			require.NoError(fullRepeater.ProcessEpochVote(*v))
		}

		if er := generator.store.GetFullEpochRecord(e); er != nil {
			ier := ier.LlrIdxFullEpochRecord{LlrFullEpochRecord: *er, Idx: e}
			require.NoError(fullRepeater.ProcessFullEpochRecord(ier))
		}
	}

	// process LLR block votes and BRs in fullReapeter
	// TODO make standalone func for that
	for b, bvs := range blockToBvsMap {
		for _, bv := range bvs {
			require.NoError(fullRepeater.ProcessBlockVotes(*bv))
		}

		if br := generator.store.GetFullBlockRecord(b); br != nil {
			ibr := ibr.LlrIdxFullBlockRecord{LlrFullBlockRecord: *br, Idx: b}
			require.NoError(fullRepeater.ProcessFullBlockRecord(ibr))
		}
	}

	fetchEvents := func() (events []*inter.EventPayload) {
		it := generator.store.table.Events.NewIterator(nil, nil)
		defer it.Release()
		for it.Next() {
			e := &inter.EventPayload{}
			if err := rlp.DecodeBytes(it.Value(), e); err != nil {
				generator.store.Log.Crit("Failed to decode event", "err", err)
			}
			if e != nil {
				events = append(events, e)
			}
		}
		return
	}

	events := fetchEvents()

	for _, e := range events {
		fullRepeater.engineMu.Lock()
		require.NoError(fullRepeater.processEvent(e))
		fullRepeater.engineMu.Unlock()
	}




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

	require.NoError(generator.store.Commit())
	require.NoError(fullRepeater.store.Commit())

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
	t.Log("Checking fullKVs <= genKVs")
	subsetOf(genKVMap, fullRepKVMap)
}

// TODO make sure there are no race conditions
// TODO plan
func TestProcessFullRecord2ThreadSafety(t *testing.T) {

}
