package integration

import (
	"strings"

	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/Fantom-foundation/lachesis-base/kvdb/leveldb"
	"github.com/Fantom-foundation/lachesis-base/kvdb/memorydb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

func dbCacheSize(name string) int {
	if name == "gossip" {
		return 64 * opt.MiB
	}
	if name == "lachesis" {
		return 4 * opt.MiB
	}
	if strings.HasPrefix(name, "lachesis-") {
		return 8 * opt.MiB
	}
	if strings.HasPrefix(name, "gossip-") {
		return 8 * opt.MiB
	}
	return 2 * opt.MiB
}

func DBProducer(dbdir string) kvdb.DBProducer {
	if dbdir == "inmemory" || dbdir == "" {
		return memorydb.NewProducer("")
	}

	return leveldb.NewProducer(dbdir, dbCacheSize)
}
