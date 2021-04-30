package evmstore

import (
	"github.com/Fantom-foundation/lachesis-base/inter/idx"

	"github.com/Fantom-foundation/go-opera/evmcore"
)

func (s *Store) GetCachedEvmBlock(n idx.Block) *evmcore.EvmBlock {
	c, ok := s.cache.EvmBlocks.Get(n)
	if !ok {
		return nil
	}
	return c.(*evmcore.EvmBlock)
}

func (s *Store) SetCachedEvmBlock(n idx.Block, b *evmcore.EvmBlock) {
	s.cache.EvmBlocks.Add(n, b, uint(b.EstimateSize()))
}
