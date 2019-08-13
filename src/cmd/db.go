package main

import (
	"os"
	"path/filepath"

	"go.etcd.io/bbolt"

	"github.com/Fantom-foundation/go-lachesis/src/logger"
)

func openDB(dir string) (db *bbolt.DB, closeDB func(), err error) {
	err = os.MkdirAll(dir, 0600)
	if err != nil {
		return
	}

	f := filepath.Join(dir, "lachesis.bolt")
	db, err = bbolt.Open(f, 0600, nil)
	if err != nil {
		return
	}

	closeDB = func() {
		if err := db.Close(); err != nil {
			logger.Get().Error(err)
		}
	}

	return
}
