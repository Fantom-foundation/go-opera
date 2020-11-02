package gossip

import (
	"errors"
	"fmt"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/inter/pos"
	"github.com/Fantom-foundation/lachesis-base/lachesis"
	"github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/go-opera/evmcore"
	"github.com/Fantom-foundation/go-opera/gossip/blockproc"
	"github.com/Fantom-foundation/go-opera/gossip/evmstore"
	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/inter/sfctype"
	"github.com/Fantom-foundation/go-opera/opera"
)

// GenesisMismatchError is raised when trying to overwrite an existing
// genesis block with an incompatible one.
type GenesisMismatchError struct {
	Stored, New hash.Event
}

// Error implements error interface.
func (e *GenesisMismatchError) Error() string {
	return fmt.Sprintf("database contains incompatible gossip genesis (have %s, new %s)", e.Stored.FullID(), e.New.FullID())
}

// ApplyGenesis writes initial state.
func (s *Store) ApplyGenesis(blockProc BlockProc, net *opera.Config) (genesisAtropos hash.Event, new bool, err error) {
	storedGenesis := s.GetBlock(0)
	if storedGenesis != nil {
		newHash := calcGenesisHash(blockProc, net)
		if storedGenesis.Atropos != newHash {
			return genesisAtropos, true, &GenesisMismatchError{storedGenesis.Atropos, newHash}
		}

		genesisAtropos = storedGenesis.Atropos
		return genesisAtropos, false, nil
	}
	// if we'here, then it's first time genesis is applied
	genesisAtropos, err = s.applyEpoch1Genesis(blockProc, net)
	if err != nil {
		return genesisAtropos, true, err
	}

	return genesisAtropos, true, err
}

// calcGenesisHash calcs hash of genesis state.
func calcGenesisHash(blockProc BlockProc, net *opera.Config) hash.Event {
	s := NewMemStore()
	defer s.Close()

	h, _ := s.applyEpoch1Genesis(blockProc, net)

	return h
}

func (s *Store) applyEpoch0Genesis(net *opera.Config) (evmBlock *evmcore.EvmBlock, err error) {
	// apply app genesis
	evmBlock, err = s.evm.ApplyGenesis(net)
	if err != nil {
		return evmBlock, err
	}

	s.SetBlockState(blockproc.BlockState{
		LastBlock:             0,
		EpochBlocks:           0,
		ValidatorStates:       make([]blockproc.ValidatorBlockState, 0),
		NextValidatorProfiles: make(map[idx.ValidatorID]sfctype.SfcValidator),
	})
	s.SetEpochState(blockproc.EpochState{
		Epoch:             0,
		EpochStart:        net.Genesis.Time - 1,
		PrevEpochStart:    net.Genesis.Time - 2,
		Validators:        pos.NewBuilder().Build(),
		ValidatorStates:   make([]blockproc.ValidatorEpochState, 0),
		ValidatorProfiles: make(map[idx.ValidatorID]sfctype.SfcValidator),
	})

	return evmBlock, nil
}

func (s *Store) applyEpoch1Genesis(blockProc BlockProc, net *opera.Config) (genesisAtropos hash.Event, err error) {
	evmBlock0, err := s.applyEpoch0Genesis(net)
	if err != nil {
		return genesisAtropos, err
	}

	evmStateReader := &EvmStateReader{store: s}
	statedb := s.evm.StateDB(hash.Hash(evmBlock0.Root))

	bs, es := s.GetBlockState(), s.GetEpochState()

	blockCtx := blockproc.BlockCtx{
		Idx:  bs.LastBlock,
		Time: net.Genesis.Time,
		CBlock: lachesis.Block{
			Atropos:  hash.Event{},
			Cheaters: make(lachesis.Cheaters, 0),
		},
	}

	sealer := blockProc.SealerModule.Start(blockCtx, bs, es)
	sealing := true
	txListener := blockProc.TxListenerModule.Start(blockCtx, bs, es, statedb)
	evmProcessor := blockProc.EVMModule.Start(blockCtx, statedb, evmStateReader, txListener.OnNewLog)

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
			return genesisAtropos, errors.New("genesis transaction reverted")
		}
	}
	if len(skippedTxs) != 0 {
		return genesisAtropos, errors.New("genesis transaction is skipped")
	}
	bs = txListener.Finalize()

	s.SetBlockState(bs)

	prettyHash := func(root common.Hash, net *opera.Config) hash.Event {
		e := inter.MutableEventPayload{}
		// for nice-looking ID
		e.SetEpoch(0)
		e.SetLamport(idx.Lamport(net.Dag.MaxEpochBlocks))
		// actual data hashed
		e.SetExtra(append(root[:], net.Genesis.ExtraData...))
		e.SetCreationTime(net.Genesis.Time)

		return e.Build().ID()
	}
	genesisAtropos = prettyHash(evmBlock.Root, net)

	block := &inter.Block{
		Time:       net.Genesis.Time,
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

	return genesisAtropos, nil
}
