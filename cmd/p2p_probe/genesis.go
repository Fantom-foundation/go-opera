package main

import (
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/go-opera/gossip"
	"github.com/Fantom-foundation/go-opera/inter/iblockproc"
	"github.com/Fantom-foundation/go-opera/inter/ier"
	"github.com/Fantom-foundation/go-opera/opera"
	"github.com/Fantom-foundation/go-opera/opera/genesisstore"
)

// LoadGenesis like gossip/Store.ApplyGenesis()
func (b *ProbeBackend) LoadGenesis(gPath string) {
	f, err := os.Open(gPath)
	if err != nil {
		panic(fmt.Errorf("Failed to open genesis file: %v", err))
	}
	defer f.Close()

	gStore, _, err := genesisstore.OpenGenesisStore(f)
	if err != nil {
		panic(fmt.Errorf("Failed to read genesis file: %v", err))
	}
	defer gStore.Close()

	var (
		g       = gStore.Genesis()
		hh      []opera.UpgradeHeight
		firstBS *iblockproc.BlockState
		firstES *iblockproc.EpochState
		lastES  *iblockproc.EpochState
	)
	g.Epochs.ForEach(func(er ier.LlrIdxFullEpochRecord) bool {
		es, bs := er.EpochState, er.BlockState

		if es.Rules.NetworkID != g.NetworkID || es.Rules.Name != g.NetworkName {
			panic("network ID/name mismatch")
		}

		if lastES == nil || es.Rules.Upgrades != lastES.Rules.Upgrades {
			hh = append(hh,
				opera.UpgradeHeight{
					Upgrades: es.Rules.Upgrades,
					Height:   bs.LastBlock.Idx + 1,
				})
		}
		lastES = &es
		if firstES == nil {
			firstES = &es
		}
		if firstBS == nil {
			firstBS = &bs
		}

		return true
	})

	if firstES == nil || firstBS == nil {
		panic("no ERs in genesis")
	}

	b.NodeInfo = &gossip.NodeInfo{
		Network:     g.NetworkID,
		Genesis:     common.Hash(g.GenesisID),
		Epoch:       firstES.Epoch,
		NumOfBlocks: firstBS.LastBlock.Idx,
	}
	b.Opera = &firstES.Rules
	b.Chain = firstES.Rules.EvmChainConfig(hh)
}
