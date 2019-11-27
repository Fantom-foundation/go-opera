package main

import (
	"errors"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

var errUnknownTx = errors.New("unknown tx")

type txsDict struct {
	txs map[common.Hash]time.Time
	sync.Mutex
}

func newTxsDict() *txsDict {
	return &txsDict{
		txs: make(map[common.Hash]time.Time),
	}
}

func (d *txsDict) Start(tx common.Hash) {
	d.Lock()
	d.txs[tx] = time.Now()
	d.Unlock()
}

func (d *txsDict) Finish(tx common.Hash) (latency time.Duration, err error) {
	d.Lock()
	defer d.Unlock()

	start, ok := d.txs[tx]
	if !ok {
		err = errUnknownTx
		return
	}
	delete(d.txs, tx)

	latency = time.Since(start)
	return
}
