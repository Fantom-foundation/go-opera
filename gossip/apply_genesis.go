package gossip

import (
	"errors"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/inter/pos"
	"github.com/Fantom-foundation/lachesis-base/lachesis"
	"github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/go-opera/evmcore"
	"github.com/Fantom-foundation/go-opera/gossip/blockproc"
	"github.com/Fantom-foundation/go-opera/gossip/evmstore"
	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/inter/drivertype"
	"github.com/Fantom-foundation/go-opera/opera"
	"github.com/Fantom-foundation/go-opera/opera/genesis"
)

// ApplyGenesis writes initial state.
func (s *Store) ApplyGenesis(blockProc BlockProc, g opera.Genesis) (genesisHash hash.Hash, err error) {
	// if we'here, then it's first time genesis is applied
	err = s.applyEpoch1Genesis(blockProc, g)
	if err != nil {
		return genesisHash, err
	}
	genesisHash = g.Hash()
	s.SetGenesisHash(genesisHash)

	return genesisHash, err
}

func (s *Store) applyEpoch0Genesis(g opera.Genesis) (evmBlock *evmcore.EvmBlock, err error) {
	// apply app genesis
	evmBlock, err = s.evm.ApplyGenesis(g)
	if err != nil {
		return evmBlock, err
	}

	// write genesis blocks
	var highestBlock idx.Block
	g.Blocks.ForEach(func(index idx.Block, block genesis.Block) {
		txHashes := make([]common.Hash, len(block.Txs))
		internalTxHashes := make([]common.Hash, len(block.InternalTxs))
		for i, tx := range block.Txs {
			txHashes[i] = tx.Hash()
		}
		for i, tx := range block.InternalTxs {
			internalTxHashes[i] = tx.Hash()
		}
		for i, tx := range append(block.InternalTxs, block.Txs...) {
			s.evm.SetTxPosition(tx.Hash(), evmstore.TxPosition{
				Block:       index,
				BlockOffset: uint32(i),
			})
			s.evm.SetTx(tx.Hash(), tx)
		}
		gasUsed := uint64(0)
		if len(block.Receipts) != 0 {
			gasUsed = block.Receipts[len(block.Receipts)-1].CumulativeGasUsed
			s.evm.SetRawReceipts(index, block.Receipts)
		}

		s.SetBlock(index, &inter.Block{
			Time:        block.Time,
			Atropos:     block.Atropos,
			Events:      hash.Events{},
			Txs:         txHashes,
			InternalTxs: internalTxHashes,
			SkippedTxs:  []uint32{},
			GasUsed:     gasUsed,
			Root:        block.Root,
		})
		s.SetBlockIndex(block.Atropos, index)
		highestBlock = index
	})

	s.SetBlockState(blockproc.BlockState{
		LastBlock:             highestBlock,
		LastStateRoot:         hash.Hash(evmBlock.Root),
		EpochBlocks:           0,
		ValidatorStates:       make([]blockproc.ValidatorBlockState, 0),
		NextValidatorProfiles: make(map[idx.ValidatorID]drivertype.Validator),
		DirtyRules:            g.Rules,
	})
	s.SetEpochState(blockproc.EpochState{
		Epoch:             g.FirstEpoch - 1,
		EpochStart:        g.PrevEpochTime,
		PrevEpochStart:    g.PrevEpochTime - 1,
		Validators:        pos.NewBuilder().Build(),
		ValidatorStates:   make([]blockproc.ValidatorEpochState, 0),
		ValidatorProfiles: make(map[idx.ValidatorID]drivertype.Validator),
		Rules:             g.Rules,
	})

	return evmBlock, nil
}

func (s *Store) applyEpoch1Genesis(blockProc BlockProc, g opera.Genesis) (err error) {
	evmBlock0, err := s.applyEpoch0Genesis(g)
	if err != nil {
		return err
	}

	evmStateReader := &EvmStateReader{store: s}
	statedb, err := s.evm.StateDB(hash.Hash(evmBlock0.Root))
	if err != nil {
		return err
	}

	bs, es := s.GetBlockState(), s.GetEpochState()
	bs.LastBlock++
	bs.EpochBlocks++

	blockCtx := blockproc.BlockCtx{
		Idx:  bs.LastBlock,
		Time: g.Time,
		CBlock: lachesis.Block{
			Atropos:  hash.Event{},
			Cheaters: make(lachesis.Cheaters, 0),
		},
	}

	sealer := blockProc.SealerModule.Start(blockCtx, bs, es)
	sealing := true
	txListener := blockProc.TxListenerModule.Start(blockCtx, bs, es, statedb)
	evmProcessor := blockProc.EVMModule.Start(blockCtx, statedb, evmStateReader, txListener.OnNewLog, es.Rules)

	// Execute genesis-internal transactions
	genesisInternalTxs := blockProc.GenesisTxTransactor.PopInternalTxs(blockCtx, bs, es, sealing, statedb)
	evmProcessor.Execute(genesisInternalTxs, true)
	bs = txListener.Finalize()

	// Execute pre-internal transactions
	preInternalTxs := blockProc.PreTxTransactor.PopInternalTxs(blockCtx, bs, es, sealing, statedb)
	evmProcessor.Execute(preInternalTxs, true)
	bs = txListener.Finalize()

	// Seal epoch if requested
	if sealing {
		sealer.Update(bs, es)
		bs, es = sealer.SealEpoch()
		s.SetEpochState(es)
		txListener.Update(bs, es)
	}

	// Execute post-internal transactions
	internalTxs := blockProc.PostTxTransactor.PopInternalTxs(blockCtx, bs, es, sealing, statedb)
	evmProcessor.Execute(internalTxs, true)
	evmBlock, skippedTxs, receipts := evmProcessor.Finalize()
	for _, r := range receipts {
		if r.Status == 0 {
			return errors.New("genesis transaction reverted")
		}
	}
	if len(skippedTxs) != 0 {
		return errors.New("genesis transaction is skipped")
	}
	bs = txListener.Finalize()
	bs.LastStateRoot = hash.Hash(evmBlock.Root)

	s.SetBlockState(bs)

	prettyHash := func(root common.Hash, g opera.Genesis) hash.Event {
		e := inter.MutableEventPayload{}
		// for nice-looking ID
		e.SetEpoch(g.FirstEpoch - 1)
		e.SetLamport(idx.Lamport(g.Rules.Dag.MaxEpochBlocks))
		// actual data hashed
		e.SetExtra(append(root[:], g.ExtraData...))
		e.SetCreationTime(g.Time)

		return e.Build().ID()
	}
	genesisAtropos := prettyHash(evmBlock.Root, g)

	block := &inter.Block{
		Time:       blockCtx.Time,
		Atropos:    genesisAtropos,
		Events:     hash.Events{},
		SkippedTxs: skippedTxs,
		GasUsed:    evmBlock.GasUsed,
		Root:       hash.Hash(evmBlock.Root),
	}

	// store txs index
	for i, tx := range append(genesisInternalTxs, append(preInternalTxs, internalTxs...)...) {
		block.InternalTxs = append(block.InternalTxs, tx.Hash())
		s.evm.SetTx(tx.Hash(), tx)
		s.evm.SetTxPosition(tx.Hash(), evmstore.TxPosition{
			Block:       blockCtx.Idx,
			BlockOffset: uint32(i),
		})
	}
	s.evm.SetReceipts(blockCtx.Idx, receipts)

	s.SetBlock(blockCtx.Idx, block)
	s.SetBlockIndex(genesisAtropos, blockCtx.Idx)
	s.SetGenesisBlockIndex(blockCtx.Idx)

	return nil
}
