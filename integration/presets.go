package integration

import (
	"github.com/Fantom-foundation/lachesis-base/kvdb/multidb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

var DefaultDBsConfig = Ldb1DBsConfig

/*
 * pbl-1 config
 */

func Pbl1DBsConfig(scale func(uint64) uint64, fdlimit uint64) DBsConfig {
	return DBsConfig{
		Routing:      Pbl1RoutingConfig(),
		RuntimeCache: Pbl1RuntimeDBsCacheConfig(scale, fdlimit),
		GenesisCache: Pbl1GenesisDBsCacheConfig(scale, fdlimit),
	}
}

func Pbl1RoutingConfig() RoutingConfig {
	return RoutingConfig{
		Table: map[string]multidb.Route{
			"": {
				Type: "pebble-fsh",
			},
			"lachesis": {
				Type:  "pebble-fsh",
				Name:  "main",
				Table: "C",
			},
			"gossip": {
				Type: "pebble-fsh",
				Name: "main",
			},
			"evm": {
				Type: "pebble-fsh",
				Name: "main",
			},
			"gossip/e": {
				Type: "pebble-fsh",
				Name: "events",
			},
			"evm/M": {
				Type: "pebble-drc",
				Name: "evm-data",
			},
			"evm-logs": {
				Type: "pebble-fsh",
				Name: "evm-logs",
			},
			"gossip-%d": {
				Type:  "leveldb-fsh",
				Name:  "epoch-%d",
				Table: "G",
			},
			"lachesis-%d": {
				Type:   "leveldb-fsh",
				Name:   "epoch-%d",
				Table:  "L",
				NoDrop: true,
			},
		},
	}
}

func Pbl1RuntimeDBsCacheConfig(scale func(uint64) uint64, fdlimit uint64) DBsCacheConfig {
	return DBsCacheConfig{
		Table: map[string]DBCacheConfig{
			"evm-data": {
				Cache:   scale(242 * opt.MiB),
				Fdlimit: fdlimit*242/700 + 1,
			},
			"evm-logs": {
				Cache:   scale(110 * opt.MiB),
				Fdlimit: fdlimit*110/700 + 1,
			},
			"main": {
				Cache:   scale(186 * opt.MiB),
				Fdlimit: fdlimit*186/700 + 1,
			},
			"events": {
				Cache:   scale(87 * opt.MiB),
				Fdlimit: fdlimit*87/700 + 1,
			},
			"epoch-%d": {
				Cache:   scale(75 * opt.MiB),
				Fdlimit: fdlimit*75/700 + 1,
			},
			"": {
				Cache:   64 * opt.MiB,
				Fdlimit: fdlimit/100 + 1,
			},
		},
	}
}

func Pbl1GenesisDBsCacheConfig(scale func(uint64) uint64, fdlimit uint64) DBsCacheConfig {
	return DBsCacheConfig{
		Table: map[string]DBCacheConfig{
			"main": {
				Cache:   scale(1024 * opt.MiB),
				Fdlimit: fdlimit*1024/3072 + 1,
			},
			"evm-data": {
				Cache:   scale(1024 * opt.MiB),
				Fdlimit: fdlimit*1024/3072 + 1,
			},
			"evm-logs": {
				Cache:   scale(1024 * opt.MiB),
				Fdlimit: fdlimit*1024/3072 + 1,
			},
			"events": {
				Cache:   scale(1 * opt.MiB),
				Fdlimit: fdlimit*1/3072 + 1,
			},
			"epoch-%d": {
				Cache:   scale(1 * opt.MiB),
				Fdlimit: fdlimit*1/3072 + 1,
			},
			"": {
				Cache:   16 * opt.MiB,
				Fdlimit: fdlimit/100 + 1,
			},
		},
	}
}

/*
 * ldb-1 config
 */

func Ldb1DBsConfig(scale func(uint64) uint64, fdlimit uint64) DBsConfig {
	return DBsConfig{
		Routing:      Ldb1RoutingConfig(),
		RuntimeCache: Ldb1RuntimeDBsCacheConfig(scale, fdlimit),
		GenesisCache: Ldb1GenesisDBsCacheConfig(scale, fdlimit),
	}
}

func Ldb1RoutingConfig() RoutingConfig {
	return RoutingConfig{
		Table: map[string]multidb.Route{
			"": {
				Type: "leveldb-fsh",
			},
			"lachesis": {
				Type:  "leveldb-fsh",
				Name:  "main",
				Table: "C",
			},
			"gossip": {
				Type: "leveldb-fsh",
				Name: "main",
			},
			"evm": {
				Type: "leveldb-fsh",
				Name: "main",
			},
			"evm-logs": {
				Type:  "leveldb-fsh",
				Name:  "main",
				Table: "L",
			},
			"gossip-%d": {
				Type:  "leveldb-fsh",
				Name:  "epoch-%d",
				Table: "G",
			},
			"lachesis-%d": {
				Type:   "leveldb-fsh",
				Name:   "epoch-%d",
				Table:  "L",
				NoDrop: true,
			},
		},
	}
}

func Ldb1RuntimeDBsCacheConfig(scale func(uint64) uint64, fdlimit uint64) DBsCacheConfig {
	return DBsCacheConfig{
		Table: map[string]DBCacheConfig{
			"main": {
				Cache:   scale(625 * opt.MiB),
				Fdlimit: fdlimit*625/700 + 1,
			},
			"epoch-%d": {
				Cache:   scale(75 * opt.MiB),
				Fdlimit: fdlimit*75/700 + 1,
			},
			"": {
				Cache:   64 * opt.MiB,
				Fdlimit: fdlimit/100 + 1,
			},
		},
	}
}

func Ldb1GenesisDBsCacheConfig(scale func(uint64) uint64, fdlimit uint64) DBsCacheConfig {
	return DBsCacheConfig{
		Table: map[string]DBCacheConfig{
			"main": {
				Cache:   scale(3072 * opt.MiB),
				Fdlimit: fdlimit,
			},
			"epoch-%d": {
				Cache:   scale(1 * opt.MiB),
				Fdlimit: fdlimit*1/3072 + 1,
			},
			"": {
				Cache:   16 * opt.MiB,
				Fdlimit: fdlimit/100 + 1,
			},
		},
	}
}

/*
 * legacy-ldb config
 */

func LdbLegacyDBsConfig(scale func(uint64) uint64, fdlimit uint64) DBsConfig {
	return DBsConfig{
		Routing:      LdbLegacyRoutingConfig(),
		RuntimeCache: LdbLegacyRuntimeDBsCacheConfig(scale, fdlimit),
		GenesisCache: LdbLegacyGenesisDBsCacheConfig(scale, fdlimit),
	}
}

func LdbLegacyRoutingConfig() RoutingConfig {
	return RoutingConfig{
		Table: map[string]multidb.Route{
			"": {
				Type: "leveldb-fsh",
			},
			"lachesis": {
				Type: "leveldb-fsh",
				Name: "lachesis",
			},
			"gossip": {
				Type: "leveldb-fsh",
				Name: "main",
			},

			"gossip/S": {
				Type:  "leveldb-fsh",
				Name:  "main",
				Table: "!",
			},
			"gossip/R": {
				Type:  "leveldb-fsh",
				Name:  "main",
				Table: "@",
			},
			"gossip/Q": {
				Type:  "leveldb-fsh",
				Name:  "main",
				Table: "#",
			},

			"gossip/T": {
				Type:  "leveldb-fsh",
				Name:  "main",
				Table: "$",
			},
			"gossip/J": {
				Type:  "leveldb-fsh",
				Name:  "main",
				Table: "%",
			},
			"gossip/E": {
				Type:  "leveldb-fsh",
				Name:  "main",
				Table: "^",
			},

			"gossip/I": {
				Type:  "leveldb-fsh",
				Name:  "main",
				Table: "&",
			},
			"gossip/G": {
				Type:  "leveldb-fsh",
				Name:  "main",
				Table: "*",
			},
			"gossip/F": {
				Type:  "leveldb-fsh",
				Name:  "main",
				Table: "(",
			},

			"evm": {
				Type: "leveldb-fsh",
				Name: "main",
			},
			"evm-logs": {
				Type:  "leveldb-fsh",
				Name:  "main",
				Table: "L",
			},
			"gossip-%d": {
				Type: "leveldb-fsh",
				Name: "gossip-%d",
			},
			"lachesis-%d": {
				Type: "leveldb-fsh",
				Name: "lachesis-%d",
			},
		},
	}
}

func LdbLegacyRuntimeDBsCacheConfig(scale func(uint64) uint64, fdlimit uint64) DBsCacheConfig {
	return DBsCacheConfig{
		Table: map[string]DBCacheConfig{
			"main": {
				Cache:   scale(564 * opt.MiB),
				Fdlimit: fdlimit*564/700 + 1,
			},
			"lachesis": {
				Cache:   scale(8 * opt.MiB),
				Fdlimit: fdlimit*8/700 + 1,
			},
			"gossip-%d": {
				Cache:   scale(64 * opt.MiB),
				Fdlimit: fdlimit*64/700 + 1,
			},
			"lachesis-%d": {
				Cache:   scale(64 * opt.MiB),
				Fdlimit: fdlimit*64/700 + 1,
			},
			"": {
				Cache:   64 * opt.MiB,
				Fdlimit: fdlimit/100 + 1,
			},
		},
	}
}

func LdbLegacyGenesisDBsCacheConfig(scale func(uint64) uint64, fdlimit uint64) DBsCacheConfig {
	return DBsCacheConfig{
		Table: map[string]DBCacheConfig{
			"main": {
				Cache:   scale(3072 * opt.MiB),
				Fdlimit: fdlimit*3072 + 1,
			},
			"lachesis": {
				Cache:   scale(1 * opt.MiB),
				Fdlimit: fdlimit*1/3072 + 1,
			},
			"gossip-%d": {
				Cache:   scale(1 * opt.MiB),
				Fdlimit: fdlimit*1/3072 + 1,
			},
			"lachesis-%d": {
				Cache:   scale(1 * opt.MiB),
				Fdlimit: fdlimit*1/3072 + 1,
			},
			"": {
				Cache:   16 * opt.MiB,
				Fdlimit: fdlimit/100 + 1,
			},
		},
	}
}

/*
 * legacy-pbl config
 */

func PblLegacyDBsConfig(scale func(uint64) uint64, fdlimit uint64) DBsConfig {
	return DBsConfig{
		Routing:      PblLegacyRoutingConfig(),
		RuntimeCache: PblLegacyRuntimeDBsCacheConfig(scale, fdlimit),
		GenesisCache: PblLegacyGenesisDBsCacheConfig(scale, fdlimit),
	}
}

func PblLegacyRoutingConfig() RoutingConfig {
	return RoutingConfig{
		Table: map[string]multidb.Route{
			"": {
				Type: "pebble-fsh",
			},
			"lachesis": {
				Type: "pebble-fsh",
				Name: "lachesis",
			},
			"gossip": {
				Type: "pebble-fsh",
				Name: "main",
			},

			"gossip/S": {
				Type:  "pebble-fsh",
				Name:  "main",
				Table: "!",
			},
			"gossip/R": {
				Type:  "pebble-fsh",
				Name:  "main",
				Table: "@",
			},
			"gossip/Q": {
				Type:  "pebble-fsh",
				Name:  "main",
				Table: "#",
			},

			"gossip/T": {
				Type:  "pebble-fsh",
				Name:  "main",
				Table: "$",
			},
			"gossip/J": {
				Type:  "pebble-fsh",
				Name:  "main",
				Table: "%",
			},
			"gossip/E": {
				Type:  "pebble-fsh",
				Name:  "main",
				Table: "^",
			},

			"gossip/I": {
				Type:  "pebble-fsh",
				Name:  "main",
				Table: "&",
			},
			"gossip/G": {
				Type:  "pebble-fsh",
				Name:  "main",
				Table: "*",
			},
			"gossip/F": {
				Type:  "pebble-fsh",
				Name:  "main",
				Table: "(",
			},

			"evm": {
				Type: "pebble-fsh",
				Name: "main",
			},
			"evm-logs": {
				Type:  "pebble-fsh",
				Name:  "main",
				Table: "L",
			},
			"gossip-%d": {
				Type: "pebble-fsh",
				Name: "gossip-%d",
			},
			"lachesis-%d": {
				Type: "pebble-fsh",
				Name: "lachesis-%d",
			},
		},
	}
}

func PblLegacyRuntimeDBsCacheConfig(scale func(uint64) uint64, fdlimit uint64) DBsCacheConfig {
	return DBsCacheConfig{
		Table: map[string]DBCacheConfig{
			"main": {
				Cache:   scale(564 * opt.MiB),
				Fdlimit: fdlimit*564/700 + 1,
			},
			"lachesis": {
				Cache:   scale(8 * opt.MiB),
				Fdlimit: fdlimit*8/700 + 1,
			},
			"gossip-%d": {
				Cache:   scale(64 * opt.MiB),
				Fdlimit: fdlimit*64/700 + 1,
			},
			"lachesis-%d": {
				Cache:   scale(64 * opt.MiB),
				Fdlimit: fdlimit*64/700 + 1,
			},
			"": {
				Cache:   64 * opt.MiB,
				Fdlimit: fdlimit/100 + 1,
			},
		},
	}
}

func PblLegacyGenesisDBsCacheConfig(scale func(uint64) uint64, fdlimit uint64) DBsCacheConfig {
	return DBsCacheConfig{
		Table: map[string]DBCacheConfig{
			"main": {
				Cache:   scale(3072 * opt.MiB),
				Fdlimit: fdlimit*3072 + 1,
			},
			"lachesis": {
				Cache:   scale(1 * opt.MiB),
				Fdlimit: fdlimit*1/3072 + 1,
			},
			"gossip-%d": {
				Cache:   scale(1 * opt.MiB),
				Fdlimit: fdlimit*1/3072 + 1,
			},
			"lachesis-%d": {
				Cache:   scale(1 * opt.MiB),
				Fdlimit: fdlimit*1/3072 + 1,
			},
			"": {
				Cache:   16 * opt.MiB,
				Fdlimit: fdlimit/100 + 1,
			},
		},
	}
}
