package gossip

import (
	"github.com/Fantom-foundation/lachesis-base/common/bigendian"
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/inter/pos"
	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/Fantom-foundation/lachesis-base/utils/simplewlru"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/inter/ibr"
	"github.com/Fantom-foundation/go-opera/inter/ier"
	"github.com/Fantom-foundation/go-opera/utils/bitmap"
)

func (s *Store) SetBlockVotes(bvs inter.LlrSignedBlockVotes) {
	s.rlp.Set(s.table.LlrBlockVotes, append(bvs.Val.Epoch.Bytes(), append(bvs.Val.LastBlock().Bytes(), bvs.Signed.Locator.ID().Bytes()...)...), &bvs)
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

func (s *Store) getLlrVoteWeight(cache *VotesCache, reader kvdb.Reader, cKey VotesCacheID, key []byte) (pos.Weight, bitmap.Set) {
	if cached := cache.Get(cKey); cached != nil {
		return cached.weight, cached.set
	}
	weightB, err := reader.Get(key)
	if err != nil {
		s.Log.Crit("Failed to get key-value", "err", err)
	}
	if weightB == nil {
		return 0, nil
	}
	weight, set := pos.Weight(bigendian.BytesToUint32(weightB[:4])), weightB[4:]
	cache.Add(cKey, VotesCacheValue{
		weight:  weight,
		set:     set,
		mutated: false,
	})
	return weight, set
}

func (s *Store) flushLlrVoteWeight(table kvdb.Writer, key []byte, weight pos.Weight, set bitmap.Set) {
	err := table.Put(key, append(bigendian.Uint32ToBytes(uint32(weight)), set...))
	if err != nil {
		s.Log.Crit("Failed to put key-value", "err", err)
	}
}

func (s *Store) flushLlrBlockVoteWeight(cKey VotesCacheID, value VotesCacheValue) {
	key := append(cKey.Block.Bytes(), append(cKey.Epoch.Bytes(), cKey.V[:]...)...)
	s.flushLlrVoteWeight(s.table.LlrBlockVotesIndex, key, value.weight, value.set)
}

func (s *Store) addLlrVoteWeight(cache *VotesCache, reader kvdb.Reader, cKey VotesCacheID, key []byte, validator idx.Validator, validatorsNum idx.Validator, diff pos.Weight) pos.Weight {
	weight, set := s.getLlrVoteWeight(cache, reader, cKey, key)
	if set != nil && set.Has(int(validator)) {
		// don't count the vote if validator already voted
		return weight
	}
	if set == nil {
		set = bitmap.New(int(validatorsNum))
	}
	set.Put(int(validator))
	weight += diff
	// save to cache which will be later flushed to the DB
	cache.Add(cKey, VotesCacheValue{
		weight:  weight,
		set:     set,
		mutated: true,
	})
	return weight
}

func (s *Store) AddLlrBlockVoteWeight(block idx.Block, epoch idx.Epoch, bv hash.Hash, val idx.Validator, vals idx.Validator, diff pos.Weight) pos.Weight {
	key := append(block.Bytes(), append(epoch.Bytes(), bv[:]...)...)
	cKey := VotesCacheID{
		Block: block,
		Epoch: epoch,
		V:     bv,
	}
	return s.addLlrVoteWeight(s.cache.LlrBlockVotesIndex, s.table.LlrBlockVotesIndex, cKey, key, val, vals, diff)
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

type VotesCacheID struct {
	Block idx.Block
	Epoch idx.Epoch
	V     hash.Hash
}

type VotesCacheValue struct {
	weight  pos.Weight
	set     bitmap.Set
	mutated bool
}

type VotesCache struct {
	votes *simplewlru.Cache
}

func NewVotesCache(maxSize int, evictedFn func(VotesCacheID, VotesCacheValue)) *VotesCache {
	votes, _ := simplewlru.NewWithEvict(uint(maxSize), maxSize, func(key interface{}, _value interface{}) {
		value := _value.(*VotesCacheValue)
		if value.mutated {
			evictedFn(key.(VotesCacheID), *value)
		}
	})
	return &VotesCache{
		votes: votes,
	}
}

func (c *VotesCache) FlushMutated(write func(VotesCacheID, VotesCacheValue)) {
	keys := c.votes.Keys()
	for _, k := range keys {
		val_, _ := c.votes.Peek(k)
		val := val_.(*VotesCacheValue)
		if val.mutated {
			write(k.(VotesCacheID), *val)
			val.mutated = false
		}
	}
}

func (c *VotesCache) Get(key VotesCacheID) *VotesCacheValue {
	if v, ok := c.votes.Get(key); ok {
		return v.(*VotesCacheValue)
	}
	return nil
}

func (c *VotesCache) Add(key VotesCacheID, val VotesCacheValue) {
	c.votes.Add(key, &val, nominalSize)
}
