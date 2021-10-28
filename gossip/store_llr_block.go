package gossip

import (
	"errors"

	"github.com/Fantom-foundation/lachesis-base/common/bigendian"
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/inter/pos"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/Fantom-foundation/go-opera/eventcheck"
	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/inter/ibr"
	"github.com/Fantom-foundation/go-opera/inter/ier"
)

func (s *Store) SetBlockVotes(bvs inter.LlrSignedBlockVotes) {
	s.rlp.Set(s.table.LlrBlockVotes, append(bvs.Epoch.Bytes(), append(bvs.LastBlock().Bytes(), bvs.EventLocator.ID().Bytes()...)...), &bvs)
}

func (s *Store) HasBlockVotes(epoch idx.Epoch, lastBlock idx.Block, id hash.Event) bool {
	ok, _ := s.table.LlrBlockVotes.Has(append(epoch.Bytes(), append(lastBlock.Bytes(), id.Bytes()...)...))
	return ok
}

func (s *Store) IterateOverlappingBlockVotesRLP(start []byte, f func(key []byte, bvs rlp.RawValue) bool) {
	it := s.table.LlrBlockVotes.NewIterator(nil, start)
	defer it.Release()
	for it.Next() {
		if !f(it.Key(), it.Value()) {
			break
		}
	}
}

func (s *Store) GetLlrBlockVoteWeight(block idx.Block, bv hash.Hash) pos.Weight {
	weightB, err := s.table.LlrBlockVotesIndex.Get(append(block.Bytes(), bv[:]...))
	if err != nil {
		s.Log.Crit("Failed to get key-value", "err", err)
	}
	if weightB == nil {
		return 0
	}
	return pos.Weight(bigendian.BytesToUint32(weightB))
}

func (s *Store) AddLlrBlockVoteWeight(block idx.Block, bv hash.Hash, diff pos.Weight) pos.Weight {
	weight := s.GetLlrBlockVoteWeight(block, bv)
	weight += diff
	err := s.table.LlrBlockVotesIndex.Put(append(block.Bytes(), bv[:]...), bigendian.Uint32ToBytes(uint32(weight)))
	if err != nil {
		s.Log.Crit("Failed to put key-value", "err", err)
	}
	return weight
}

func (s *Store) SetLlrBlockResult(block idx.Block, bv hash.Hash) {
	err := s.table.LlrBlockResults.Put(block.Bytes(), bv.Bytes())
	if err != nil {
		s.Log.Crit("Failed to put key-value", "err", err)
	}
}

func (s *Store) GetLlrBlockResult(block idx.Block) *hash.Hash {
	bvB, err := s.table.LlrBlockResults.Get(block.Bytes())
	if err != nil {
		s.Log.Crit("Failed to get key-value", "err", err)
	}
	if bvB == nil {
		return nil
	}
	bv := hash.BytesToHash(bvB)
	return &bv
}

func actualizeLowestIndex(current, upd uint64, exists func(uint64) bool) uint64 {
	if current == upd {
		current++
		for exists(current) {
			current++
		}
	}
	return current
}

func (s *Service) ProcessBlockVote(block idx.Block, weight pos.Weight, totalWeight pos.Weight, bv hash.Hash, llrs *LlrState) {
	newWeight := s.store.AddLlrBlockVoteWeight(block, bv, weight)
	if newWeight >= totalWeight/3+1 {
		wonBr := s.store.GetLlrBlockResult(block)
		if wonBr == nil {
			s.store.SetLlrBlockResult(block, bv)
			llrs.LowestBlockToDecide = idx.Block(actualizeLowestIndex(uint64(llrs.LowestBlockToDecide), uint64(block), func(u uint64) bool {
				return s.store.GetLlrBlockResult(idx.Block(u)) != nil
			}))
		} else if *wonBr != bv {
			s.Log.Error("LLR voting doublesign is met", "block", block)
		}
	}
}

func (s *Service) ProcessBlockVotes(bvs inter.LlrSignedBlockVotes) error {
	// engineMu should be locked here
	if len(bvs.Votes) == 0 {
		// short circuit if no records
		return nil
	}
	if s.store.HasBlockVotes(bvs.Epoch, bvs.LastBlock(), bvs.EventLocator.ID()) {
		return eventcheck.ErrAlreadyProcessedBVs
	}
	done := s.procLogger.BlockVotesConnectionStarted(bvs)
	defer done()
	vid := bvs.EventLocator.Creator
	// get the validators group
	epoch := bvs.LlrBlockVotes.Epoch
	_, es := s.store.GetHistoryBlockEpochState(epoch)
	if es == nil {
		return eventcheck.ErrUnknownEpochBVs
	}

	llrs := s.store.GetLlrState()
	b := bvs.Start
	for _, bv := range bvs.Votes {
		s.ProcessBlockVote(b, es.Validators.Get(vid), es.Validators.TotalWeight(), bv, &llrs)
		b++
	}
	s.store.SetLlrState(llrs)
	s.store.SetBlockVotes(bvs)
	lBVs := s.store.GetLastBVs()
	lBVs.Lock()
	if bvs.LastBlock() > lBVs.Val[vid] {
		lBVs.Val[vid] = bvs.LastBlock()
		s.store.SetLastBVs(lBVs)
	}
	lBVs.Unlock()

	return nil
}

func (s *Service) ProcessFullBlockRecord(br ibr.LlrIdxFullBlockRecord) error {
	// engineMu should NOT be locked here
	if s.store.HasBlock(br.Idx) {
		return eventcheck.ErrAlreadyProcessedBR
	}
	done := s.procLogger.BlockRecordConnectionStarted(br)
	defer done()
	res := s.store.GetLlrBlockResult(br.Idx)
	if res == nil {
		return eventcheck.ErrUndecidedBR
	}

	if br.Hash() != *res {
		return errors.New("block record hash mismatch")
	}

	txHashes := make([]common.Hash, 0, len(br.Txs))
	internalTxHashes := make([]common.Hash, 0, 2)
	for _, tx := range br.Txs {
		if tx.GasPrice().Sign() == 0 {
			internalTxHashes = append(internalTxHashes, tx.Hash())
		} else {
			txHashes = append(txHashes, tx.Hash())
		}
		s.store.EvmStore().SetTx(tx.Hash(), tx)
	}

	s.store.EvmStore().SetRawReceipts(br.Idx, br.Receipts)
	for _, tx := range br.Txs {
		s.store.EvmStore().SetTx(tx.Hash(), tx)
	}
	s.store.SetBlock(br.Idx, &inter.Block{
		Time:        br.Time,
		Atropos:     br.Atropos,
		Events:      hash.Events{},
		Txs:         txHashes,
		InternalTxs: internalTxHashes,
		SkippedTxs:  []uint32{},
		GasUsed:     br.GasUsed,
		Root:        br.Root,
	})
	s.store.SetBlockIndex(br.Atropos, br.Idx)
	s.engineMu.Lock()
	defer s.engineMu.Unlock()
	updateLowestBlockToFill(br.Idx, s.store)

	return nil
}

func updateLowestBlockToFill(block idx.Block, store *Store) {
	llrs := store.GetLlrState()
	llrs.LowestBlockToFill = idx.Block(actualizeLowestIndex(uint64(llrs.LowestBlockToFill), uint64(block), func(u uint64) bool {
		return store.GetBlock(idx.Block(u)) != nil
	}))
	store.SetLlrState(llrs)
}

type LlrState struct {
	LowestEpochToDecide idx.Epoch
	LowestEpochToFill   idx.Epoch

	LowestBlockToDecide idx.Block
	LowestBlockToFill   idx.Block
}

func (s *Store) SetLlrState(llrs LlrState) {
	s.rlp.Set(s.table.LlrState, []byte{}, &llrs)
}

func (s *Store) GetLlrState() LlrState {
	w, _ := s.rlp.Get(s.table.LlrState, []byte{}, &LlrState{}).(*LlrState)
	if w == nil {
		return LlrState{}
	}
	return *w
}

func (s *Store) GetFullBlockRecord(n idx.Block) *ibr.LlrFullBlockRecord {
	block := s.GetBlock(n)
	if block == nil {
		return nil
	}
	txs := s.GetBlockTxs(n, block)
	receipts, _ := s.EvmStore().GetRawReceipts(n)
	if receipts == nil {
		receipts = []*types.ReceiptForStorage{}
	}
	return &ibr.LlrFullBlockRecord{
		Atropos:  block.Atropos,
		Root:     block.Root,
		Txs:      txs,
		Receipts: receipts,
		Time:     block.Time,
		GasUsed:  block.GasUsed,
	}
}

func (s *Store) GetFullEpochRecord(epoch idx.Epoch) *ier.LlrFullEpochRecord {
	hbs, hes := s.GetHistoryBlockEpochState(epoch)
	if hbs == nil || hes == nil {
		return nil
	}
	return &ier.LlrFullEpochRecord{
		BlockState: *hbs,
		EpochState: *hes,
	}
}

type LlrFullBlockRecordRLP struct {
	Atropos     hash.Event
	Root        hash.Hash
	Txs         types.Transactions
	ReceiptsRLP rlp.RawValue
	Time        inter.Timestamp
	GasUsed     uint64
}

type LlrIdxFullBlockRecordRLP struct {
	LlrFullBlockRecordRLP
	Idx idx.Block
}

var emptyReceiptsRLP, _ = rlp.EncodeToBytes([]*types.ReceiptForStorage{})

func (s *Store) IterateFullBlockRecordsRLP(start idx.Block, f func(b idx.Block, br rlp.RawValue) bool) {
	it := s.table.Blocks.NewIterator(nil, start.Bytes())
	defer it.Release()
	for it.Next() {
		block := &inter.Block{}
		err := rlp.DecodeBytes(it.Value(), block)
		if err != nil {
			s.Log.Crit("Failed to decode block", "err", err)
		}
		n := idx.BytesToBlock(it.Key())
		txs := s.GetBlockTxs(n, block)
		receiptsRLP := s.EvmStore().GetRawReceiptsRLP(n)
		if receiptsRLP == nil {
			receiptsRLP = emptyReceiptsRLP
		}
		br := LlrIdxFullBlockRecordRLP{
			LlrFullBlockRecordRLP: LlrFullBlockRecordRLP{
				Atropos:     block.Atropos,
				Root:        block.Root,
				Txs:         txs,
				ReceiptsRLP: receiptsRLP,
				Time:        block.Time,
				GasUsed:     block.GasUsed,
			},
			Idx: n,
		}
		encoded, err := rlp.EncodeToBytes(br)
		if err != nil {
			s.Log.Crit("Failed to encode BR", "err", err)
		}

		if !f(n, encoded) {
			break
		}
	}
}
