package gossip

import (
	"errors"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/kvdb/batched"
	"github.com/syndtr/goleveldb/leveldb/opt"

	"github.com/Fantom-foundation/go-opera/inter/iblockproc"
	"github.com/Fantom-foundation/go-opera/inter/ibr"
	"github.com/Fantom-foundation/go-opera/inter/ier"
	"github.com/Fantom-foundation/go-opera/opera/genesis"
	"github.com/Fantom-foundation/go-opera/utils/dbutil/autocompact"
)

// ApplyGenesis writes initial state.
func (s *Store) ApplyGenesis(g genesis.Genesis) (genesisHash hash.Hash, err error) {
	// use batching wrapper for hot tables
	unwrap := s.WrapTablesAsBatched()
	defer unwrap()

	// write epochs
	var topEr *ier.LlrIdxFullEpochRecord
	g.Epochs.ForEach(func(er ier.LlrIdxFullEpochRecord) bool {
		if er.EpochState.Rules.NetworkID != g.NetworkID || er.EpochState.Rules.Name != g.NetworkName {
			err = errors.New("network ID/name mismatch")
			return false
		}
		if topEr == nil {
			topEr = &er
		}
		s.WriteFullEpochRecord(er)
		return true
	})
	if err != nil {
		return genesisHash, err
	}
	if topEr == nil {
		return genesisHash, errors.New("no ERs in genesis")
	}
	var prevEs *iblockproc.EpochState
	s.ForEachHistoryBlockEpochState(func(bs iblockproc.BlockState, es iblockproc.EpochState) bool {
		s.WriteUpgradeHeight(bs, es, prevEs)
		prevEs = &es
		return true
	})
	s.SetBlockEpochState(topEr.BlockState, topEr.EpochState)
	s.FlushBlockEpochState()

	// write blocks
	g.Blocks.ForEach(func(br ibr.LlrIdxFullBlockRecord) bool {
		s.WriteFullBlockRecord(br)
		return true
	})

	// write EVM items
	err = s.evm.ApplyGenesis(g)
	if err != nil {
		return genesisHash, err
	}

	// write LLR state
	s.setLlrState(LlrState{
		LowestEpochToDecide: topEr.Idx + 1,
		LowestEpochToFill:   topEr.Idx + 1,
		LowestBlockToDecide: topEr.BlockState.LastBlock.Idx + 1,
		LowestBlockToFill:   topEr.BlockState.LastBlock.Idx + 1,
	})
	s.FlushLlrState()

	s.SetGenesisID(g.GenesisID)
	s.SetGenesisBlockIndex(topEr.BlockState.LastBlock.Idx)

	return genesisHash, err
}

func (s *Store) WrapTablesAsBatched() (unwrap func()) {
	origTables := s.table

	batchedBlocks := batched.Wrap(autocompact.Wrap2M(s.table.Blocks, opt.GiB, 16*opt.GiB, false, "blocks"))
	s.table.Blocks = batchedBlocks

	batchedBlockHashes := batched.Wrap(s.table.BlockHashes)
	s.table.BlockHashes = batchedBlockHashes

	unwrapEVM := s.evm.WrapTablesAsBatched()
	return func() {
		unwrapEVM()
		_ = batchedBlocks.Flush()
		_ = batchedBlockHashes.Flush()
		s.table = origTables
	}
}
