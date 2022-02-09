package gossip

import (
	//"errors"
	"fmt"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/inter/pos"
	"github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/go-opera/eventcheck"
	"github.com/Fantom-foundation/go-opera/gossip/evmstore"
	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/inter/ibr"
	"github.com/Fantom-foundation/go-opera/inter/ier"
)

func actualizeLowestIndex(current, upd uint64, exists func(uint64) bool) uint64 {
	if current == upd {
		current++
		for exists(current) {
			current++
		}
	}
	return current
}

func (s *Service) processBlockVote(block idx.Block, epoch idx.Epoch, bv hash.Hash, val idx.Validator, vals *pos.Validators, llrs *LlrState) {

	fmt.Println("processBlockVote vals", vals)
	fmt.Println("processBlockVote val", val)
	fmt.Println("processBlockVote vals.GetWeightByIdx(val)", vals.GetWeightByIdx(val))
	newWeight := s.store.AddLlrBlockVoteWeight(block, epoch, bv, val, vals.Len(), vals.GetWeightByIdx(val))

	if newWeight >= vals.TotalWeight()/3+1 {
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

func (s *Service) processBlockVotes(bvs inter.LlrSignedBlockVotes) error {
	// engineMu should be locked here
	if len(bvs.Val.Votes) == 0 {
		// short circuit if no records
		return nil
	}
	if s.store.HasBlockVotes(bvs.Val.Epoch, bvs.Val.LastBlock(), bvs.Signed.Locator.ID()) {
		return eventcheck.ErrAlreadyProcessedBVs
	}
	done := s.procLogger.BlockVotesConnectionStarted(bvs)
	defer done()
	vid := bvs.Signed.Locator.Creator

	// get the validators group
	epoch := bvs.Signed.Locator.Epoch
	fmt.Println("processBlockVotes epoch", epoch)
	// TODO make sure it is correct decision to call GetHistoryEpochState(epoch-1)
	es := s.store.GetHistoryEpochState(epoch)
	if es == nil {
		fmt.Println("processBlockVotes es == nil")
		return eventcheck.ErrUnknownEpochBVs
	}

	fmt.Println("processBlockVotes vid", vid)
	fmt.Println("processBlockVotes es.Validators", es.Validators.String())

	s.store.ModifyLlrState(func(llrs *LlrState) {
		b := bvs.Val.Start
		for _, bv := range bvs.Val.Votes {
			fmt.Println("processBlockVotes es.Validators.GetIdx(vid)", es.Validators.GetIdx(vid))
			s.processBlockVote(b, bvs.Val.Epoch, bv, es.Validators.GetIdx(vid), es.Validators, llrs)
			b++
		}
	})
	s.store.SetBlockVotes(bvs)
	lBVs := s.store.GetLastBVs()
	lBVs.Lock()
	if bvs.Val.LastBlock() > lBVs.Val[vid] {
		lBVs.Val[vid] = bvs.Val.LastBlock()
		s.store.SetLastBVs(lBVs)
	}
	lBVs.Unlock()

	return nil
}

func (s *Service) ProcessBlockVotes(bvs inter.LlrSignedBlockVotes) error {
	s.engineMu.Lock()
	defer s.engineMu.Unlock()
	err := s.processBlockVotes(bvs)
	if err == nil {
		s.mayCommit(false)
	}
	return err
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
		fmt.Println("ProcessFullBlockRecord GetLlrBlockResult res == nil")
		return eventcheck.ErrUndecidedBR
	}

	/* TODO resolve it
	if br.Hash() != *res {
		return errors.New("block record hash mismatch")
	}
	*/

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

	if len(br.Receipts) != 0 {
		// Note: it's possible for receipts to get indexed twice by BR and block processing
		indexRawReceipts(s.store, br.Receipts, br.Txs, br.Idx, br.Atropos)
	}
	for i, tx := range br.Txs {
		s.store.EvmStore().SetTx(tx.Hash(), tx)
		s.store.EvmStore().SetTxPosition(tx.Hash(), evmstore.TxPosition{
			Block:       br.Idx,
			BlockOffset: uint32(i),
		})
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
	if s.verWatcher != nil {
		// Note: it's possible for logs to get indexed twice by BR and block processing
		for _, r := range br.Receipts {
			for _, l := range r.Logs {
				s.verWatcher.OnNewLog(l)
			}
		}
	}
	updateLowestBlockToFill(br.Idx, s.store)
	s.mayCommit(false)

	return nil
}

func (s *Service) processRawEpochVote(epoch idx.Epoch, ev hash.Hash, val idx.Validator, vals *pos.Validators, llrs *LlrState) {
	newWeight := s.store.AddLlrEpochVoteWeight(epoch, ev, val, vals.Len(), vals.GetWeightByIdx(val))
	fmt.Println("processRawEpochVote epoch, newWeight, val, ev", epoch, newWeight, val, ev)
	fmt.Println("processRawEpochVote vals.TotalWeight()/3+1", vals.TotalWeight()/3+1)
	if newWeight >= vals.TotalWeight()/3+1 {
		fmt.Println("processRawEpochVote newWeight >= vals.TotalWeight()/3+1")
		wonEr := s.store.GetLlrEpochResult(epoch)
		if wonEr == nil {
			fmt.Println("processRawEpochVote wonEr == nil")
			s.store.SetLlrEpochResult(epoch, ev)
			llrs.LowestEpochToDecide = idx.Epoch(actualizeLowestIndex(uint64(llrs.LowestEpochToDecide), uint64(epoch), func(u uint64) bool {
				return s.store.GetLlrEpochResult(idx.Epoch(u)) != nil
			}))
		} else if *wonEr != ev {
			s.Log.Error("LLR voting doublesign is met", "epoch", epoch)
		}
	}
}

func (s *Service) processEpochVote(ev inter.LlrSignedEpochVote) error {
	// engineMu should be locked here
	if ev.Val.Epoch == 0 {
		// short circuit if no records
		return nil
	}
	if s.store.HasEpochVote(ev.Val.Epoch, ev.Signed.Locator.ID()) {
		return eventcheck.ErrAlreadyProcessedEV
	}
	done := s.procLogger.EpochVoteConnectionStarted(ev)
	defer done()
	vid := ev.Signed.Locator.Creator
	fmt.Println("processEpochVote vid", vid)

	// get the validators group
	es := s.store.GetHistoryEpochState(ev.Val.Epoch - 1)
	if es == nil {
		return eventcheck.ErrUnknownEpochEV
	}
	fmt.Println("processEpochVote es.Validators.GetIdx(vid)", es.Validators.GetIdx(vid))

	s.store.ModifyLlrState(func(llrs *LlrState) {
		s.processRawEpochVote(ev.Val.Epoch, ev.Val.Vote, es.Validators.GetIdx(vid), es.Validators, llrs)
	})
	s.store.SetEpochVote(ev)
	lEVs := s.store.GetLastEVs()
	lEVs.Lock()
	if ev.Val.Epoch > lEVs.Val[vid] {
		lEVs.Val[vid] = ev.Val.Epoch
		s.store.SetLastEVs(lEVs)
	}
	lEVs.Unlock()

	return nil
}

func (s *Service) ProcessEpochVote(ev inter.LlrSignedEpochVote) error {
	s.engineMu.Lock()
	defer s.engineMu.Unlock()
	err := s.processEpochVote(ev)
	if err == nil {
		s.mayCommit(false)
	}
	return err
}

func (s *Service) ProcessFullEpochRecord(er ier.LlrIdxFullEpochRecord) error {
	// engineMu should NOT be locked here
	if s.store.HasHistoryBlockEpochState(er.Idx) {
		return eventcheck.ErrAlreadyProcessedER
	}
	done := s.procLogger.EpochRecordConnectionStarted(er)
	defer done()

	res := s.store.GetLlrEpochResult(er.Idx)
	if res == nil {
		return eventcheck.ErrUndecidedER
	}
	/* it looks like this is a bug
	fmt.Println("ProcessFullEpochRecord er.Idx", er.Idx)
	fmt.Println("ProcessFullEpochRecord res", res.String())
	fmt.Println("ProcessFullEpochRecord ev.Vote = hash.HexToHash(0x12)", hash.HexToHash("0x12").String())
	fmt.Println("ProcessFullEpochRecord er.Hash", er.Hash().String())
	fmt.Println("ProcessFullEpochRecord er.Blockstate", er.BlockState)
	fmt.Println("ProcessFullEpochRecord er.Epochstate", er.EpochState)

	ev = hash.HexToHash("0x12") this is epoch vote from fakeEvent()
	epoch = 2, ev = 0x0000000000000000000000000000000000000000000000000000000000000012
	s.store.SetLlrEpochResult(epoch, ev) in processRawEpochVote
	res is ev

	er.Hash() is
	func (er LlrFullEpochRecord) Hash() hash.Hash {
	return hash.Of(er.BlockState.Hash().Bytes(), er.EpochState.Hash().Bytes())
	}

	if er.Hash() != *res {
		return errors.New("epoch record hash mismatch")
	}
	*/

	s.store.SetHistoryBlockEpochState(er.Idx, er.BlockState, er.EpochState)
	s.store.SetEpochBlock(er.BlockState.LastBlock.Idx+1, er.Idx)
	s.engineMu.Lock()
	defer s.engineMu.Unlock()
	updateLowestEpochToFill(er.Idx, s.store)
	s.mayCommit(false)

	return nil
}

func updateLowestBlockToFill(block idx.Block, store *Store) {
	store.ModifyLlrState(func(llrs *LlrState) {
		llrs.LowestBlockToFill = idx.Block(actualizeLowestIndex(uint64(llrs.LowestBlockToFill), uint64(block), func(u uint64) bool {
			return store.GetBlock(idx.Block(u)) != nil
		}))
	})
}

func updateLowestEpochToFill(epoch idx.Epoch, store *Store) {
	store.ModifyLlrState(func(llrs *LlrState) {
		llrs.LowestEpochToFill = idx.Epoch(actualizeLowestIndex(uint64(llrs.LowestEpochToFill), uint64(epoch), func(u uint64) bool {
			return store.HasHistoryBlockEpochState(idx.Epoch(u))
		}))
	})
}
