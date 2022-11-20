package erigon

import (
	"context"
	"io"

	"github.com/Fantom-foundation/lachesis-base/common/bigendian"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ledgerwatch/erigon-lib/kv"
)

var emptyCode = crypto.Keccak256(nil)

// Write iterates over erigon kv.PlainState records and populates io.Writer
// TODO iterate over not only plainstate as well as other tables kv.Code etc
func Write(writer io.Writer, genesisKV kv.RwDB) (accounts int, err error) {

	tx, err := genesisKV.BeginRo(context.Background())
	if err != nil {
		return 0, err
	}

	defer tx.Rollback()

	c, err := tx.Cursor(kv.PlainState)
	if err != nil {
		return accounts, err
	}
	defer c.Close()

	for k, v, e := c.First(); k != nil; k, v, e = c.Next() {
		if e != nil {
			return accounts, e
		}

		_, err := writer.Write(bigendian.Uint32ToBytes(uint32(len(k))))
		if err != nil {
			return accounts, err
		}
		_, err = writer.Write(k)
		if err != nil {
			return accounts, err
		}
		_, err = writer.Write(bigendian.Uint32ToBytes(uint32(len(v))))
		if err != nil {
			return accounts, err
		}
		_, err = writer.Write(v)
		if err != nil {
			return accounts, err
		}
		accounts++
	}

	log.Info("Erigon write", "Plainstate Iterations count", accounts)
	return accounts, nil
}
