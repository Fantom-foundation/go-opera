package erigon

import (
	"encoding/binary"
	"fmt"
	"path/filepath"
	"runtime"
	"time"

	"context"

	libcommon "github.com/ledgerwatch/erigon-lib/common"
	"github.com/ledgerwatch/erigon-lib/etl"
	"github.com/ledgerwatch/erigon-lib/kv"

	"github.com/ledgerwatch/erigon-lib/common/length"
	"github.com/ledgerwatch/erigon/common"
	"github.com/ledgerwatch/erigon/common/dbutils"
	"github.com/ledgerwatch/erigon/turbo/trie"
	"github.com/ledgerwatch/log/v3"
)

// GenerateHashState extracts data from kv.Plainstate and writes to kv.HashedAccounts and kv.HashedStorage
func GenerateHashedState(logPrefix string, db kv.RwDB, ctx context.Context) error {
	tmpDir := filepath.Join(DefaultDataDir(), "erigon", "hashedstate")
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

	return tx.Commit()
}

// TODO write tests
// readPlainStateOnce reads kv.Plainstate and then loads data into kv.HashedAccounts and kv.HashedStorage
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

// not a stable method to generate HashState
func GenerateHashState2(tx kv.RwTx) error {
	// address tx.Rollback() in upper level

	c, err := tx.Cursor(kv.PlainState)
	if err != nil {
		return err
	}
	h := common.NewHasher()

	for k, v, err := c.First(); k != nil; k, v, err = c.Next() {
		if err != nil {
			return fmt.Errorf("interate over plain state: %w", err)
		}
		var newK []byte
		if len(k) == common.AddressLength {
			newK = make([]byte, common.HashLength)
		} else {
			newK = make([]byte, common.HashLength*2+common.IncarnationLength)
		}
		h.Sha.Reset()                         //?
		h.Sha.Write(k[:common.AddressLength]) //?
		h.Sha.Read(newK[:common.HashLength])  //?
		if len(k) > common.AddressLength {
			copy(newK[common.HashLength:], k[common.AddressLength:common.AddressLength+common.IncarnationLength])
			h.Sha.Reset()
			h.Sha.Write(k[common.AddressLength+common.IncarnationLength:])
			h.Sha.Read(newK[common.HashLength+common.IncarnationLength:])
			if err = tx.Put(kv.HashedStorage, newK, common.CopyBytes(v)); err != nil {
				return fmt.Errorf("insert hashed key: %w", err)
			}
		} else {
			if err = tx.Put(kv.HashedAccounts, newK, common.CopyBytes(v)); err != nil {
				return fmt.Errorf("insert hashed key: %w", err)
			}
		}
	}
	c.Close()

	/*
		if err := tx.Commit(); err != nil {
			return err
		}
	*/

	return nil
}

// not stable method to generate trie root
func CalcTrieRoot2(db kv.RwDB) (common.Hash, error) {
	tx, err := db.BeginRw(context.Background())
	if err != nil {
		return common.Hash{}, err
	}
	defer tx.Rollback()

	root, err := trie.CalcRoot("", tx)
	if err != nil {
		return common.Hash{}, err
	}

	return root, nil
}

// GenerateHashedStatePut does the same thing as GenerateHashedStateLoad but in a different manner using tx.Put method. It iterates over kv.Plainstate records and fill in kv.HashedAccounts and kv.HashedStorage records.
func GenerateHashedStatePut(tx kv.RwTx) error {

	c, err := tx.Cursor(kv.PlainState)
	if err != nil {
		return err
	}
	h := common.NewHasher()

	for k, v, err := c.First(); k != nil; k, v, err = c.Next() {
		if err != nil {
			return fmt.Errorf("interate over plain state: %w", err)
		}
		var newK []byte
		if len(k) == common.AddressLength {
			newK = make([]byte, common.HashLength)
		} else {
			newK = make([]byte, common.HashLength*2+common.IncarnationLength)
		}
		h.Sha.Reset()                         //?
		h.Sha.Write(k[:common.AddressLength]) //?
		h.Sha.Read(newK[:common.HashLength])  //?
		if len(k) > common.AddressLength {
			copy(newK[common.HashLength:], k[common.AddressLength:common.AddressLength+common.IncarnationLength])
			h.Sha.Reset()
			h.Sha.Write(k[common.AddressLength+common.IncarnationLength:])
			h.Sha.Read(newK[common.HashLength+common.IncarnationLength:])
			if err = tx.Put(kv.HashedStorage, newK, common.CopyBytes(v)); err != nil {
				return fmt.Errorf("insert hashed key: %w", err)
			}
		} else {
			if err = tx.Put(kv.HashedAccounts, newK, common.CopyBytes(v)); err != nil {
				return fmt.Errorf("insert hashed key: %w", err)
			}
		}
	}
	c.Close()

	/*
		if err := tx.Commit(); err != nil {
			return err
		}
	*/

	return nil
}
