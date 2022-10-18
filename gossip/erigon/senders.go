package erigon

import (
	"context"
	"fmt"

	"github.com/ledgerwatch/erigon-lib/kv"

	"github.com/ledgerwatch/erigon/common"

	ecommon "github.com/ledgerwatch/erigon/common"
)

func WriteSenders(tx kv.Putter, preimages map[common.Hash][]byte) error {
	if preimages == nil {
		return fmt.Errorf("preimages map is nil")
	}

	for key, value := range preimages {
		if err := tx.Put(kv.Senders, key.Bytes(), value); err != nil {
			return fmt.Errorf("failed to store preimages: %w", err)
		}
	}

	return nil
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
