package integration

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/Fantom-foundation/lachesis-base/abft"
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/dag"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/Fantom-foundation/lachesis-base/kvdb/flushable"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"

	"github.com/Fantom-foundation/go-opera/gossip"
	"github.com/Fantom-foundation/go-opera/gossip/blockproc/eventmodule"
	"github.com/Fantom-foundation/go-opera/gossip/blockproc/evmmodule"
	"github.com/Fantom-foundation/go-opera/gossip/blockproc/sealmodule"
	"github.com/Fantom-foundation/go-opera/gossip/blockproc/sfcmodule"
	"github.com/Fantom-foundation/go-opera/opera"
	"github.com/Fantom-foundation/go-opera/utils/adapters/vecmt2dagidx"
	"github.com/Fantom-foundation/go-opera/vecmt"
)

type Configs struct {
	Opera         gossip.Config
	OperaStore    gossip.StoreConfig
	Lachesis      abft.Config
	LachesisStore abft.StoreConfig
	VectorClock   vecmt.IndexConfig
}

func panics(name string) func(error) {
	return func(err error) {
		log.Crit(fmt.Sprintf("%s error", name), "err", err)
	}
}

type GossipStoreAdapter struct {
	*gossip.Store
}

func (g *GossipStoreAdapter) GetEvent(id hash.Event) dag.Event {
	e := g.Store.GetEvent(id)
	if e == nil {
		return nil
	}
	return e
}

// MakeEngine makes consensus engine from config.
func MakeEngine(dbs *flushable.SyncedPool, cfg Configs, g opera.Genesis) (*abft.Lachesis, *vecmt.Index, *gossip.Store, gossip.BlockProc) {
	gdb := gossip.NewStore(dbs, cfg.OperaStore)

	cMainDb, err := dbs.OpenDB("lachesis")
	if err != nil {
		utils.Fatalf("Failed to open Lachesis DB: %v", err)
	}

	cGetEpochDB := func(epoch idx.Epoch) kvdb.DropableStore {
		db, err := dbs.OpenDB(fmt.Sprintf("lachesis-%d", epoch))
		if err != nil {
			utils.Fatalf("Failed to open Lachesis-%d DB: %v", epoch, err)
		}
		return db
	}
	cdb := abft.NewStore(cMainDb, cGetEpochDB, panics("Lachesis store"), cfg.LachesisStore)

	// write genesis

	blockProc := gossip.BlockProc{
		SealerModule:        sealmodule.New(g.Rules),
		TxListenerModule:    sfcmodule.NewSfcTxListenerModule(g.Rules),
		GenesisTxTransactor: sfcmodule.NewSfcTxGenesisTransactor(g),
		PreTxTransactor:     sfcmodule.NewSfcTxPreTransactor(g.Rules),
		PostTxTransactor:    sfcmodule.NewSfcTxTransactor(g.Rules),
		EventsModule:        eventmodule.New(g.Rules),
		EVMModule:           evmmodule.New(g.Rules),
	}

	err = gdb.Migrate()
	if err != nil {
		utils.Fatalf("Failed to migrate Gossip DB: %v", err)
	}
	genesisHash, isNew, err := gdb.ApplyGenesis(blockProc, g)
	if err != nil {
		utils.Fatalf("Failed to write Gossip genesis state: %v", err)
	}

	if isNew {
		err = cdb.ApplyGenesis(&abft.Genesis{
			Epoch:      gdb.GetEpoch(),
			Validators: gdb.GetValidators(),
		})
		if err != nil {
			utils.Fatalf("Failed to write Lachesis genesis state: %v", err)
		}
	}

	err = dbs.Flush(genesisHash.Bytes())
	if err != nil {
		utils.Fatalf("Failed to flush genesis state: %v", err)
	}

	if isNew {
		log.Info("Applied genesis state", "hash", genesisHash.String())
	} else {
		log.Info("Genesis state is already written", "hash", genesisHash.String())
	}

	// create consensus
	vecClock := vecmt.NewIndex(panics("Vector clock"), cfg.VectorClock)
	engine := abft.NewLachesis(cdb, &GossipStoreAdapter{gdb}, vecmt2dagidx.Wrap(vecClock), panics("Lachesis"), cfg.Lachesis)

	return engine, vecClock, gdb, blockProc
}

// SetAccountKey sets key into accounts manager and unlocks it with pswd.
func SetAccountKey(
	am *accounts.Manager, key *ecdsa.PrivateKey, pswd string,
) (
	acc accounts.Account,
) {
	kss := am.Backends(keystore.KeyStoreType)
	if len(kss) < 1 {
		log.Crit("Keystore is not found")
		return
	}
	ks := kss[0].(*keystore.KeyStore)

	acc = accounts.Account{
		Address: crypto.PubkeyToAddress(key.PublicKey),
	}

	imported, err := ks.ImportECDSA(key, pswd)
	if err == nil {
		acc = imported
	} else if err.Error() != "account already exists" {
		log.Crit("Failed to import key", "err", err)
	}

	err = ks.Unlock(acc, pswd)
	if err != nil {
		log.Crit("failed to unlock key", "err", err)
	}

	return
}
