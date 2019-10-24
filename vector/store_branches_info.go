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
	key := []byte("current")

	vi.setRlp(vi.table.BranchesInfo, key, info)

	vi.cache.BranchesInfo.Add(string(key), *info)
}

func (vi *Index) getBranchesInfo() *branchesInfo {
	key := []byte("current")

	wInterface, okGet := vi.cache.BranchesInfo.Get(string(key))
	wVal, okType := wInterface.(branchesInfo)
	if okGet && okType {
		return &wVal
	}

	w, exists := vi.getRlp(vi.table.BranchesInfo, key, &branchesInfo{}).(*branchesInfo)
	if !exists {
		return nil
	}

	vi.cache.BranchesInfo.Add(key, *w)

	return w
}
