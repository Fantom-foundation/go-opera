package integration

import (
	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/Fantom-foundation/lachesis-base/kvdb/leveldb"
	"github.com/Fantom-foundation/lachesis-base/kvdb/memorydb"
)

func DBProducer(dbdir string) kvdb.DbProducer {
	if dbdir == "inmemory" || dbdir == "" {
		return memorydb.NewProducer("")
	}

	return leveldb.NewProducer(dbdir)
}
