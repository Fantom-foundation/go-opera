package gossip

import (
	"errors"

	"github.com/Fantom-foundation/lachesis-base/common/bigendian"
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/inter/pos"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/Fantom-foundation/go-opera/eventcheck"
	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/inter/ier"
)

const (
	maxEpochPackVotes = 10000
)

func (s *Store) SetEpochVote(ev inter.LlrSignedEpochVote) {
	s.rlp.Set(s.table.LlrEpochVotes, append(ev.Epoch.Bytes(), ev.EventLocator.ID().Bytes()...), &ev)
}

func (s *Store) HasEpochVote(epoch idx.Epoch, id hash.Event) bool {
	ok, _ := s.table.LlrEpochVotes.Has(append(epoch.Bytes(), id.Bytes()...))
	return ok
}

func (s *Store) iterateEpochVotesRLP(prefix []byte, f func(key []byte, ev rlp.RawValue) bool) {
	it := s.table.LlrEpochVotes.NewIterator(prefix, nil)
	defer it.Release()
	for it.Next() {
		if !f(it.Key(), it.Value()) {
			break
		}
	}
}

func (s *Store) GetLlrEpochVoteWeight(epoch idx.Epoch, bv hash.Hash) pos.Weight {
	weightB, err := s.table.LlrEpochVoteIndex.Get(append(epoch.Bytes(), bv[:]...))
	if err != nil {
		s.Log.Crit("Failed to get key-value", "err", err)
	}
	if weightB == nil {
		return 0
	}
	return pos.Weight(bigendian.BytesToUint32(weightB))
}

func (s *Store) AddLlrEpochVoteWeight(epoch idx.Epoch, bv hash.Hash, diff pos.Weight) pos.Weight {
	weight := s.GetLlrEpochVoteWeight(epoch, bv)
	weight += diff
	err := s.table.LlrEpochVoteIndex.Put(append(epoch.Bytes(), bv[:]...), bigendian.Uint32ToBytes(uint32(weight)))
	if err != nil {
		s.Log.Crit("Failed to put key-value", "err", err)
	}
	return weight
}

func (s *Store) SetLlrEpochResult(epoch idx.Epoch, bv hash.Hash) {
	err := s.table.LlrEpochResults.Put(epoch.Bytes(), bv.Bytes())
	if err != nil {
		s.Log.Crit("Failed to put key-value", "err", err)
	}
}

func (s *Store) GetLlrEpochResult(epoch idx.Epoch) *hash.Hash {
	bvB, err := s.table.LlrEpochResults.Get(epoch.Bytes())
	if err != nil {
		s.Log.Crit("Failed to get key-value", "err", err)
	}
	if bvB == nil {
		return nil
	}
	bv := hash.BytesToHash(bvB)
	return &bv
}

func (s *Service) processEpochVote(epoch idx.Epoch, weight pos.Weight, totalWeight pos.Weight, bv hash.Hash, llrs *LlrState) {
	newWeight := s.store.AddLlrEpochVoteWeight(epoch, bv, weight)
	if newWeight >= totalWeight/3+1 {
		wonBr := s.store.GetLlrEpochResult(epoch)
		if wonBr == nil {
			s.store.SetLlrEpochResult(epoch, bv)
			llrs.LowestEpochToDecide = idx.Epoch(actualizeLowestIndex(uint64(llrs.LowestEpochToDecide), uint64(epoch), func(u uint64) bool {
				return s.store.GetLlrEpochResult(idx.Epoch(u)) != nil
			}))
		} else if *wonBr != bv {
			s.Log.Error("LLR voting doublesign is met", "epoch", epoch)
		}
	}
}

func (s *Service) ProcessEpochVote(ev inter.LlrSignedEpochVote) error {
	// engineMu should be locked here
	if ev.Epoch == 0 {
		// short circuit if no records
		return nil
	}
	if s.store.HasEpochVote(ev.Epoch, ev.EventLocator.ID()) {
		return eventcheck.ErrAlreadyProcessedEV
	}
	done := s.procLogger.EpochVoteConnectionStarted(ev)
	defer done()
	vid := ev.EventLocator.Creator
	// get the validators group
	_, es := s.store.GetHistoryBlockEpochState(ev.LlrEpochVote.Epoch - 1)
	if es == nil {
		return eventcheck.ErrUnknownEpochEV
	}

	llrs := s.store.GetLlrState()
	s.processEpochVote(ev.Epoch, es.Validators.Get(vid), es.Validators.TotalWeight(), ev.Vote, &llrs)
	s.store.SetLlrState(llrs)
	s.store.SetEpochVote(ev)
	lEVs := s.store.GetLastEVs()
	lEVs.Lock()
	if ev.Epoch > lEVs.Val[vid] {
		lEVs.Val[vid] = ev.Epoch
		s.store.SetLastEVs(lEVs)
	}
	lEVs.Unlock()

	return nil
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

	if er.Hash() != *res {
		return errors.New("epoch record hash mismatch")
	}

	s.store.SetHistoryBlockEpochState(er.Idx, er.BlockState, er.EpochState)
	s.store.SetEpochBlock(er.BlockState.LastBlock.Idx+1, er.Idx)
	s.engineMu.Lock()
	defer s.engineMu.Unlock()
	updateLowestEpochToFill(er.Idx, s.store)

	return nil
}

func updateLowestEpochToFill(epoch idx.Epoch, store *Store) {
	llrs := store.GetLlrState()
	llrs.LowestEpochToFill = idx.Epoch(actualizeLowestIndex(uint64(llrs.LowestEpochToFill), uint64(epoch), func(u uint64) bool {
		return store.HasHistoryBlockEpochState(idx.Epoch(u))
	}))
	store.SetLlrState(llrs)
}

type LlrIdxFullEpochRecordRLP struct {
	RecordRLP rlp.RawValue
	Idx       idx.Epoch
}

type LlrEpochPackRLP struct {
	VotesRLP []rlp.RawValue
	Record   LlrIdxFullEpochRecordRLP
}

func (s *Store) IterateEpochPacksRLP(start idx.Epoch, f func(epoch idx.Epoch, ep rlp.RawValue) bool) {
	it := s.table.BlockEpochStateHistory.NewIterator(nil, start.Bytes())
	defer it.Release()
	for it.Next() {
		epoch := idx.BytesToEpoch(it.Key())

		evs := make([]rlp.RawValue, 0, 20)
		s.iterateEpochVotesRLP(it.Key(), func(key []byte, ev rlp.RawValue) bool {
			evs = append(evs, ev)
			return len(evs) < maxEpochPackVotes
		})
		if len(evs) == 0 {
			continue
		}
		ep := &LlrEpochPackRLP{
			VotesRLP: evs,
			Record: LlrIdxFullEpochRecordRLP{
				RecordRLP: it.Value(),
				Idx:       epoch,
			},
		}
		encoded, err := rlp.EncodeToBytes(ep)
		if err != nil {
			s.Log.Crit("Failed to encode BR", "err", err)
		}
		if !f(epoch, encoded) {
			break
		}
	}
}
