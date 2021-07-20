package gsignercache

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	lru "github.com/hashicorp/golang-lru"
)

var (
	globalCache, _ = lru.New(40000)
)

type WlruCache struct {
	Cache *lru.Cache
}

func (w *WlruCache) Add(txid common.Hash, c types.CachedSender) {
	w.Cache.Add(txid, c)
}

func (w *WlruCache) Get(txid common.Hash) *types.CachedSender {
	ic, ok := w.Cache.Get(txid)
	if !ok {
		return nil
	}
	c := ic.(types.CachedSender)
	return &c
}

func Wrap(signer types.Signer) types.Signer {
	return types.WrapWithCachedSigner(signer, &WlruCache{globalCache})
}
