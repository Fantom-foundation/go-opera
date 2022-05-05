package erigon

import (
	"time"
	"runtime"
	"encoding/binary"
	"fmt"

	"context"

	libcommon "github.com/ledgerwatch/erigon-lib/common"
	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/erigon-lib/etl"

	"github.com/ledgerwatch/erigon/common"
	"github.com/ledgerwatch/erigon-lib/common/length"
	"github.com/ledgerwatch/erigon/common/dbutils"
	"github.com/ledgerwatch/log/v3"
)

// SpawnHashState extracts data from kv.Plainstate and writes to kv.HashedAccounts and kv.HashedStorage
func SpawnHashState(logPrefix string, db kv.RwDB, tmpDir string, ctx context.Context) error  {
	tx, err := db.BeginRw(context.Background())
	if err != nil {
			return err
	}
	defer tx.Rollback()

	if err := readPlainStateOnce(
		logPrefix,
		tx,
		tmpDir,
		etl.IdentityLoadFunc,
		ctx.Done(),
	); err != nil {
		return err
	}


   // ?
   /*
	if err := tx.Commit(); err != nil {
		return err
	}
  */

	return nil

}

func readPlainStateOnce(
	logPrefix string,
	db kv.RwTx,
	tmpdir string,
	loadFunc etl.LoadFunc,
	quit <-chan struct{},
) error {
	bufferSize := etl.BufferOptimalSize

	accCollector := etl.NewCollector(logPrefix, tmpdir, etl.NewSortableBuffer(bufferSize))
	defer accCollector.Close()
	storageCollector := etl.NewCollector(logPrefix, tmpdir, etl.NewSortableBuffer(bufferSize))
	defer storageCollector.Close()

	t := time.Now()
	logEvery := time.NewTicker(30 * time.Second)
	defer logEvery.Stop()
	var m runtime.MemStats

	c, err := db.Cursor(kv.PlainState)
	if err != nil {
		return err
	}
	defer c.Close()

	convertAccFunc := func(key []byte) ([]byte, error) {
		hash, err := common.HashData(key)
		return hash[:], err
	}

	convertStorageFunc := func(key []byte) ([]byte, error) {
		addrHash, err := common.HashData(key[:length.Addr])
		if err != nil {
			return nil, err
		}
		inc := binary.BigEndian.Uint64(key[length.Addr:])
		secKey, err := common.HashData(key[length.Addr+length.Incarnation:])
		if err != nil {
			return nil, err
		}
		compositeKey := dbutils.GenerateCompositeStorageKey(addrHash, inc, secKey)
		return compositeKey, nil
	}

	var startkey []byte

	// reading kv.PlainState
	for k, v, e := c.Seek(startkey); k != nil; k, v, e = c.Next() {
		if e != nil {
			return e
		}
		
		if err := libcommon.Stopped(quit); err != nil {
			return err
		}
	

		if len(k) == 20 {
			newK, err := convertAccFunc(k)
			if err != nil {
				return err
			}
			if err := accCollector.Collect(newK, v); err != nil {
				return err
			}
		} else {
			newK, err := convertStorageFunc(k)
			if err != nil {
				return err
			}
			if err := storageCollector.Collect(newK, v); err != nil {
				return err
			}
		}

		select {
		default:
		case <-logEvery.C:
			runtime.ReadMemStats(&m)
			log.Info(fmt.Sprintf("[%s] ETL [1/2] Extracting", logPrefix), "current key", fmt.Sprintf("%x...", k[:6]), "alloc", libcommon.ByteCount(m.Alloc), "sys", libcommon.ByteCount(m.Sys))
		}
	}

	log.Trace(fmt.Sprintf("[%s] Extraction finished", logPrefix), "took", time.Since(t))
	defer func(t time.Time) {
		log.Trace(fmt.Sprintf("[%s] Load finished", logPrefix), "took", time.Since(t))
	}(time.Now())

	args := etl.TransformArgs{
		Quit: quit,
	}

	if err := accCollector.Load(db, kv.HashedAccounts, loadFunc, args); err != nil {
		return err
	}

	if err := storageCollector.Load(db, kv.HashedStorage, loadFunc, args); err != nil {
		return err
	}

	return nil
}