package gossip

import (
	"github.com/Fantom-foundation/lachesis-base/common/bigendian"
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/inter/pos"
	"github.com/Fantom-foundation/lachesis-base/kvdb"
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

func (s *Store) getLlrVoteWeight(reader kvdb.Reader, key []byte) (pos.Weight, bitmap.Set) {
	weightB, err := reader.Get(key)
	if err != nil {
		s.Log.Crit("Failed to get key-value", "err", err)
	}
	if weightB == nil {
		return 0, nil
	}
	return pos.Weight(bigendian.BytesToUint32(weightB[:4])), weightB[4:]
}

func (s *Store) addLlrVoteWeight(table kvdb.Store, key []byte, val idx.Validator, vals idx.Validator, diff pos.Weight) pos.Weight {
	weight, set := s.getLlrVoteWeight(table, key)
	if set != nil && set.Has(int(val)) {
		// don't count the vote if validator already voted
		return weight
	}
	if set == nil {
		set = bitmap.New(int(vals))
	}
	set.Put(int(val))
	weight += diff
	// save to the DB
	err := table.Put(key, append(bigendian.Uint32ToBytes(uint32(weight)), set...))
	if err != nil {
		s.Log.Crit("Failed to put key-value", "err", err)
	}
	return weight
}

func (s *Store) AddLlrBlockVoteWeight(block idx.Block, epoch idx.Epoch, bv hash.Hash, val idx.Validator, vals idx.Validator, diff pos.Weight) pos.Weight {
	key := append(block.Bytes(), append(epoch.Bytes(), bv[:]...)...)
	return s.addLlrVoteWeight(s.table.LlrBlockVotesIndex, key, val, vals, diff)
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
