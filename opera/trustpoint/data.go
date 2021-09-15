package trustpoint

import (
	"math/big"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/Fantom-foundation/go-opera/gossip/blockproc"
	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/opera/genesis/gpos"
)

type (
	Metadata struct {
		Validators    gpos.Validators
		FirstEpoch    idx.Epoch
		Time          inter.Timestamp
		PrevEpochTime inter.Timestamp
		ExtraData     []byte
		DriverOwner   common.Address
		TotalSupply   *big.Int
	}
)

func (s *Store) GetBlockEpochState() (*blockproc.BlockState, *blockproc.EpochState) {
	bs, ok := s.rlp.Get(s.table.BlockEpochState, []byte("b"), &blockproc.BlockState{}).(*blockproc.BlockState)
	if !ok {
		log.Crit("Block state reading failed")
	}
	es, ok := s.rlp.Get(s.table.BlockEpochState, []byte("e"), &blockproc.EpochState{}).(*blockproc.EpochState)
	if !ok {
		log.Crit("Epoch state reading failed")
	}

	return bs, es
}

func (s *Store) SetBlockEpochState(bs *blockproc.BlockState, es *blockproc.EpochState) {
	s.rlp.Set(s.table.BlockEpochState, []byte("b"), bs)
	s.rlp.Set(s.table.BlockEpochState, []byte("e"), es)
}

func (s *Store) SetEvent(e *inter.EventPayload) {
	key := e.ID().Bytes()
	s.rlp.Set(s.table.Events, key, e)
}

func (s *Store) GetEvent(id hash.Event) *inter.EventPayload {
	key := id.Bytes()
	e := s.rlp.Get(s.table.Events, key, &inter.EventPayload{}).(*inter.EventPayload)
	return e
}

func (s *Store) SetBlock(index idx.Block, block *inter.Block) {
	s.rlp.Set(s.table.Blocks, index.Bytes(), block)
}

func (s *Store) ForEachBlock(fn func(idx.Block, *inter.Block)) {
	it := s.table.Blocks.NewIterator(nil, nil)
	defer it.Release()
	for it.Next() {
		index := idx.BytesToBlock(it.Key())
		block := &inter.Block{}
		err := rlp.DecodeBytes(it.Value(), block)
		if err != nil {
			log.Crit("Trustpoint blocks error", "err", err)
		}
		fn(index, block)
	}
}

func (s *Store) SetTx(id common.Hash, tx *types.Transaction) {
	s.rlp.Set(s.table.Txs, id.Bytes(), tx)
}

func (s *Store) GetTx(id common.Hash) *types.Transaction {
	tx := s.rlp.Get(s.table.Txs, id.Bytes(), &types.Transaction{}).(*types.Transaction)
	return tx
}

func (s *Store) SetReceipts(n idx.Block, receipts types.Receipts) {
	receiptsStorage := make([]*types.ReceiptForStorage, len(receipts))
	for i, r := range receipts {
		receiptsStorage[i] = (*types.ReceiptForStorage)(r)
	}

	buf, err := rlp.EncodeToBytes(receiptsStorage)
	if err != nil {
		s.Log.Crit("Failed to encode rlp", "err", err)
	}

	if err := s.table.Receipts.Put(n.Bytes(), buf); err != nil {
		s.Log.Crit("Failed to put key-value", "err", err)
	}
}

func (s *Store) GetRawReceipts(n idx.Block) []*types.ReceiptForStorage {
	buf, err := s.table.Receipts.Get(n.Bytes())
	if err != nil {
		s.Log.Crit("Failed to get key-value", "err", err)
	}
	if buf == nil {
		return nil
	}

	var receipts []*types.ReceiptForStorage
	err = rlp.DecodeBytes(buf, &receipts)
	if err != nil {
		s.Log.Crit("Failed to decode rlp", "err", err, "size", len(buf))
	}

	return receipts
}

func (s *Store) SetRawEvmItem(key, value []byte) {
	err := s.table.RawEvmItems.Put(key, value)
	if err != nil {
		s.Log.Crit("Failed to put EVM item", "err", err)
	}
}

func (s *Store) GetRawEvmItem() kvdb.Iteratee {
	return s.table.RawEvmItems
}
