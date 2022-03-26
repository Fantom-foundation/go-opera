package integration

import (
	"fmt"

	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/Fantom-foundation/lachesis-base/kvdb/cachedproducer"
	"github.com/Fantom-foundation/lachesis-base/kvdb/flushable"
	"github.com/Fantom-foundation/lachesis-base/kvdb/multidb"
	"github.com/Fantom-foundation/lachesis-base/kvdb/skipkeys"
)

type RoutingConfig struct {
	Table map[string]multidb.Route
}

func DefaultRoutingConfig() RoutingConfig {
	return RoutingConfig{
		Table: map[string]multidb.Route{
			"": {
				Type: "pebble",
			},
			"lachesis": {
				Type:  "pebble",
				Name:  "main",
				Table: ">",
			},
			"gossip": {
				Type: "pebble",
				Name: "main",
			},
			"evm": {
				Type: "pebble",
				Name: "main",
			},
			"gossip/e": {
				Type: "pebble",
				Name: "events",
			},
			"evm/M": {
				Type: "pebble",
				Name: "evm-data",
			},
			"evm-logs": {
				Type: "pebble",
				Name: "evm-logs",
			},
			"gossip-%d": {
				Type:  "leveldb",
				Name:  "epoch-%d",
				Table: "G",
			},
			"lachesis-%d": {
				Type:  "leveldb",
				Name:  "epoch-%d",
				Table: "L",
			},
		},
	}
}

func MakeFlushableMultiProducer(rawProducers map[multidb.TypeName]kvdb.IterableDBProducer, cfg RoutingConfig) (kvdb.FullDBProducer, func(), error) {
	flushables := make(map[multidb.TypeName]kvdb.FullDBProducer)
	var flushID []byte
	var err error
	var closeDBs = func() {}
	for typ, producer := range rawProducers {
		existingDBs := producer.Names()
		flushablePool := flushable.NewSyncedPool(producer, FlushIDKey)
		prevCloseDBs := closeDBs
		closeDBs = func() {
			prevCloseDBs()
			_ = flushablePool.Close()
		}
		flushID, err = flushablePool.Initialize(existingDBs, flushID)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to open existing databases: %v", err)
		}
		flushables[typ] = cachedproducer.WrapAll(flushablePool)
	}

	p, err := makeMultiProducer(flushables, cfg)
	return p, closeDBs, err
}

func MakeRawMultiProducer(rawProducers map[multidb.TypeName]kvdb.IterableDBProducer, cfg RoutingConfig) (kvdb.FullDBProducer, error) {
	flushables := make(map[multidb.TypeName]kvdb.FullDBProducer)
	for typ, producer := range rawProducers {
		flushables[typ] = cachedproducer.WrapAll(&DummyFlushableProducer{producer})
	}

	p, err := makeMultiProducer(flushables, cfg)
	return p, err
}

func makeMultiProducer(producers map[multidb.TypeName]kvdb.FullDBProducer, cfg RoutingConfig) (kvdb.FullDBProducer, error) {
	multi, err := multidb.NewProducer(producers, cfg.Table, TablesKey)
	if err != nil {
		return nil, fmt.Errorf("failed to construct multidb: %v", err)
	}

	err = multi.Verify()
	if err != nil {
		return nil, fmt.Errorf("incompatible chainstore DB layout: %v. Try to use 'db migrate' to recover", err)
	}
	return skipkeys.WrapAllProducer(multi, MetadataPrefix), nil
}
