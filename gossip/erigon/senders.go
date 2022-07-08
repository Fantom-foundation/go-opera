package erigon

import (
	"fmt"

	"github.com/ledgerwatch/erigon-lib/kv"

	"github.com/ledgerwatch/erigon/common"

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

/*
func ReadSenderAddress(tx kv.Getter, key common.Hash) (common.Address, error) {
	val, err := tx.GetOne(kv.Senders, key.Bytes()); 
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to store block senders: %w", err)
	}
	return common.BytesToAddress(val), nil
}
*/