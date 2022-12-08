package utils

import (
	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/Fantom-foundation/lachesis-base/kvdb/table"
)

func NewTableOrSelf(db kvdb.Store, prefix []byte) kvdb.Store {
	if len(prefix) == 0 {
		return db
	}
	return table.New(db, prefix)
}
