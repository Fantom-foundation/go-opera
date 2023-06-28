package txtime

import (
	"time"

	"github.com/Fantom-foundation/lachesis-base/utils/wlru"
	"github.com/ethereum/go-ethereum/common"
)

var global, _ = wlru.New(10000, 10000)

func Add(txid common.Hash, t time.Time) {
	global.Add(txid, t, 1)
}

func Get(txid common.Hash) (time.Time, bool) {
	v, has := global.Get(txid)
	if has {
		return v.(time.Time), true
	}
	return time.Time{}, false
}
