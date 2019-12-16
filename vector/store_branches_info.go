package vector

import (
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/Fantom-foundation/go-lachesis/kvdb"
)

func (vi *Index) setRlp(table kvdb.KeyValueStore, key []byte, val interface{}) {
	buf, err := rlp.EncodeToBytes(val)
	if err != nil {
		vi.Log.Crit("Failed to encode rlp", "err", err)
	}

	if err := table.Put(key, buf); err != nil {
		vi.Log.Crit("Failed to put key-value", "err", err)
	}
}

func (vi *Index) getRlp(table kvdb.KeyValueStore, key []byte, to interface{}) interface{} {
	buf, err := table.Get(key)
	if err != nil {
		vi.Log.Crit("Failed to get key-value", "err", err)
	}
	if buf == nil {
		return nil
	}

	err = rlp.DecodeBytes(buf, to)
	if err != nil {
		vi.Log.Crit("Failed to decode rlp", "err", err, "size", len(buf))
	}
	return to
}

func (vi *Index) setBranchesInfo(info *branchesInfo) {
	key := []byte("c")

	vi.setRlp(vi.table.BranchesInfo, key, info)
}

func (vi *Index) getBranchesInfo() *branchesInfo {
	key := []byte("c")

	w, exists := vi.getRlp(vi.table.BranchesInfo, key, &branchesInfo{}).(*branchesInfo)
	if !exists {
		return nil
	}

	return w
}
