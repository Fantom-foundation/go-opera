package trustpoint

import (
	"crypto/sha256"
	"math/big"

	"github.com/Fantom-foundation/lachesis-base/common/bigendian"
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/Fantom-foundation/go-opera/gossip/blockproc"
	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/opera/genesis"
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

func (s *Store) SetBlock(index idx.Block, block genesis.Block) {
	s.rlp.Set(s.table.Blocks, index.Bytes(), &block)
}

func (s *Store) ForEachBlock(fn func(idx.Block, genesis.Block)) {
	it := s.table.Blocks.NewIterator(nil, nil)
	defer it.Release()
	for it.Next() {
		index := idx.BytesToBlock(it.Key())
		block := genesis.Block{}
		err := rlp.DecodeBytes(it.Value(), &block)
		if err != nil {
			log.Crit("Genesis blocks error", "err", err)
		}
		fn(index, block)
	}
}

func (s *Store) SetRawEvmItem(key, value []byte) {
	err := s.table.RawEvmItems.Put(key, value)
	if err != nil {
		s.Log.Crit("Failed to put key-value", "err", err)
	}
}

func (s *Store) GetRawEvmItem() kvdb.Iteratee {
	return s.table.RawEvmItems
}

func (s *Store) Hash() hash.Hash {
	hasher := sha256.New()
	it := s.db.NewIterator(nil, nil)
	defer it.Release()
	for it.Next() {
		k := it.Key()
		v := it.Value()
		hasher.Write(bigendian.Uint32ToBytes(uint32(len(k))))
		hasher.Write(k)
		hasher.Write(bigendian.Uint32ToBytes(uint32(len(v))))
		hasher.Write(v)
	}
	return hash.BytesToHash(hasher.Sum(nil))
}
