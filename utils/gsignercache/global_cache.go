package gsignercache

import (
	"github.com/Fantom-foundation/lachesis-base/utils/wlru"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

var (
	globalWlruCache, _ = wlru.New(10000, 10000)
)

type WlruCache struct {
	Cache *wlru.Cache
}

func (w *WlruCache) Add(txid common.Hash, c types.CachedSender) {
	w.Cache.Add(txid, c, 1)
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
	return types.WrapWithCachedSigner(signer, &WlruCache{globalWlruCache})
}
