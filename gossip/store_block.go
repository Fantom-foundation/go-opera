package gossip

import (
	"math"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/Fantom-foundation/go-opera/inter"
)

func (s *Store) GetGenesisID() *hash.Hash {
	valBytes, err := s.table.Genesis.Get([]byte("g"))
	if err != nil {
		s.Log.Crit("Failed to get key-value", "err", err)
	}
	if len(valBytes) == 0 {
		return nil
	}
	val := hash.BytesToHash(valBytes)
	return &val
}

func (s *Store) SetGenesisID(val hash.Hash) {
	err := s.table.Genesis.Put([]byte("g"), val.Bytes())
	if err != nil {
		s.Log.Crit("Failed to put key-value", "err", err)
	}
}

// SetBlock stores chain block.
func (s *Store) SetBlock(n idx.Block, b *inter.Block) {
	s.rlp.Set(s.table.Blocks, n.Bytes(), b)

	// Add to LRU cache.
	s.cache.Blocks.Add(n, b, uint(b.EstimateSize()))
}

// GetBlock returns stored block.
func (s *Store) GetBlock(n idx.Block) *inter.Block {
	// Get block from LRU cache first.
	if c, ok := s.cache.Blocks.Get(n); ok {
		return c.(*inter.Block)
	}

	block, _ := s.rlp.Get(s.table.Blocks, n.Bytes(), &inter.Block{}).(*inter.Block)

	// Add to LRU cache.
	if block != nil {
		s.cache.Blocks.Add(n, block, uint(block.EstimateSize()))
	}

	return block
}

func (s *Store) HasBlock(n idx.Block) bool {
	has, _ := s.table.Blocks.Has(n.Bytes())
	return has
}

func (s *Store) ForEachBlock(fn func(index idx.Block, block *inter.Block)) {
	it := s.table.Blocks.NewIterator(nil, nil)
	defer it.Release()
	for it.Next() {
		var block inter.Block
		err := rlp.DecodeBytes(it.Value(), &block)
		if err != nil {
			s.Log.Crit("Failed to decode block", "err", err)
		}
		fn(idx.BytesToBlock(it.Key()), &block)
	}
}

// SetBlockIndex stores chain block index.
func (s *Store) SetBlockIndex(id hash.Event, n idx.Block) {
	if err := s.table.BlockHashes.Put(id.Bytes(), n.Bytes()); err != nil {
		s.Log.Crit("Failed to put key-value", "err", err)
	}

	s.cache.BlockHashes.Add(id, n, nominalSize)
}

// GetBlockIndex returns stored block index.
func (s *Store) GetBlockIndex(id hash.Event) *idx.Block {
	nVal, ok := s.cache.BlockHashes.Get(id)
	if ok {
		n, ok := nVal.(idx.Block)
		if ok {
			return &n
		}
	}

	buf, err := s.table.BlockHashes.Get(id.Bytes())
	if err != nil {
		s.Log.Crit("Failed to get key-value", "err", err)
	}
	if buf == nil {
		return nil
	}
	n := idx.BytesToBlock(buf)

	s.cache.BlockHashes.Add(id, n, nominalSize)

	return &n
}

// SetGenesisBlockIndex stores genesis block index.
func (s *Store) SetGenesisBlockIndex(n idx.Block) {
	if err := s.table.Genesis.Put([]byte("i"), n.Bytes()); err != nil {
		s.Log.Crit("Failed to put key-value", "err", err)
	}
}

// GetGenesisBlockIndex returns stored genesis block index.
func (s *Store) GetGenesisBlockIndex() *idx.Block {
	buf, err := s.table.Genesis.Get([]byte("i"))
	if err != nil {
		s.Log.Crit("Failed to get key-value", "err", err)
	}
	if buf == nil {
		return nil
	}
	n := idx.BytesToBlock(buf)

	return &n
}

func (s *Store) GetGenesisTime() inter.Timestamp {
	n := s.GetGenesisBlockIndex()
	if n == nil {
		return 0
	}
	block := s.GetBlock(*n)
	if block == nil {
		return 0
	}
	return block.Time
}

func (s *Store) SetEpochBlock(b idx.Block, e idx.Epoch) {
	err := s.table.EpochBlocks.Put((math.MaxUint64 - b).Bytes(), e.Bytes())
	if err != nil {
		s.Log.Crit("Failed to set key-value", "err", err)
	}
}

func (s *Store) FindBlockEpoch(b idx.Block) idx.Epoch {
	it := s.table.EpochBlocks.NewIterator(nil, (math.MaxUint64 - b).Bytes())
	defer it.Release()
	if !it.Next() {
		return 0
	}
	return idx.BytesToEpoch(it.Value())
}

func (s *Store) GetBlockTxs(n idx.Block, block *inter.Block) types.Transactions {
	if cached := s.evm.GetCachedEvmBlock(n); cached != nil {
		return cached.Transactions
	}

	transactions := make(types.Transactions, 0, len(block.Txs)+len(block.InternalTxs)+len(block.Events)*10)
	for _, txid := range block.InternalTxs {
		tx := s.evm.GetTx(txid)
		if tx == nil {
			log.Crit("Internal tx not found", "tx", txid.String())
			continue
		}
		transactions = append(transactions, tx)
	}
	for _, txid := range block.Txs {
		tx := s.evm.GetTx(txid)
		if tx == nil {
			log.Crit("Tx not found", "tx", txid.String())
			continue
		}
		transactions = append(transactions, tx)
	}
	for _, id := range block.Events {
		e := s.GetEventPayload(id)
		if e == nil {
			log.Crit("Block event not found", "event", id.String())
			continue
		}
		transactions = append(transactions, e.Txs()...)
	}

	transactions = inter.FilterSkippedTxs(transactions, block.SkippedTxs)

	return transactions
}
