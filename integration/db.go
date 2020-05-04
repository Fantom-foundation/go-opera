package integration

import (
	"github.com/Fantom-foundation/go-lachesis/kvdb"
	"github.com/Fantom-foundation/go-lachesis/kvdb/leveldb"
	"github.com/Fantom-foundation/go-lachesis/kvdb/memorydb"
)

func dbProducer(dbdir string) kvdb.DbProducer {
	if dbdir == "inmemory" || dbdir == "" {
		return memorydb.NewProducer("")
	}

	return leveldb.NewProducer(dbdir)
}
