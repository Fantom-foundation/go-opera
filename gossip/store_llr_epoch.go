package gossip

import (
	"github.com/Fantom-foundation/lachesis-base/common/bigendian"
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/inter/pos"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/Fantom-foundation/go-opera/inter"
)

const (
	maxEpochPackVotes = 10000
)

func (s *Store) SetEpochVote(ev inter.LlrSignedEpochVote) {
	s.rlp.Set(s.table.LlrEpochVotes, append(ev.Val.Epoch.Bytes(), ev.Signed.Locator.ID().Bytes()...), &ev)
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
