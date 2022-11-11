package erigon

import (
	"compress/gzip"
	"context"
	"io"
	"os"
	"strings"
	"time"
	"path/filepath"

	"github.com/ethereum/go-ethereum/rlp"

	"github.com/ethereum/go-ethereum/log"

	"github.com/ledgerwatch/erigon-lib/kv"

	"github.com/ledgerwatch/erigon/common"
	ecommon "github.com/ledgerwatch/erigon/common"
	"github.com/ledgerwatch/erigon/crypto"
)



// WritePreimagesToSenders writes preimages to erigon kv.Senders table
func WritePreimagesToSenders(db kv.RwDB) error {

	// THis path denotes path to preimages in profiling6 server
	defaultPreimagesPath := filepath.Join(DefaultDataDir(), "/preimages/preimages.gz")
	start := time.Now()
	log.Info("Reading preimages", "from file", defaultPreimagesPath)

	// Open the file handle and potentially unwrap the gzip stream
	fh, err := os.Open(defaultPreimagesPath)
	if err != nil {
		return err
	}
	defer fh.Close()

	var reader io.Reader = fh
	if strings.HasSuffix(defaultPreimagesPath, ".gz") {
		if reader, err = gzip.NewReader(reader); err != nil {
			return err
		}
	}
	stream := rlp.NewStream(reader, 0)

	buf := newAppendBuffer(bufferOptimalSize)

	log.Info("Writing preimages to buf")
	for {
		// Read the next entry and ensure it's not junk
		var blob []byte

		if err := stream.Decode(&blob); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		// Accumulate the preimages and flush when enough ws gathered
		key, val := crypto.Keccak256Hash(blob), common.CopyBytes(blob)
		buf.Put(key.Bytes(), val)
	}

	log.Info("Reading preimages and writing preimages to buf completed", "elapsed", common.PrettyDuration(time.Since(start)))

	// Sort data in buffer
	log.Info("Buffer length", "is", buf.Len())
	log.Info("Sorting data in buffer started...")
	start = time.Now()
	// there are no duplicates keys in buf
	buf.Sort()
	log.Info("Sorting data in buf completed", "elapsed", common.PrettyDuration(time.Since(start)))
	// sorting items in buffer to write into erigon kv.Senders efficiently
	return buf.writeIntoTable(db, kv.Senders)
}

func addressFromPreimage(db kv.RwDB, accHash common.Hash) (ecommon.Address, error) {
	var addr ecommon.Address
	if err := db.View(context.Background(), func(tx kv.Tx) error {
		val, err := tx.GetOne(kv.Senders, accHash.Bytes())
		if err != nil {
			return err
		}
		addr = ecommon.BytesToAddress(val)

		return nil
	}); err != nil {
		return ecommon.Address{}, nil
	}

	return addr, nil
}
