package main

import (
	"context"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/Fantom-foundation/go-lachesis/cmd/txn-storm/meta"
	"github.com/Fantom-foundation/go-lachesis/logger"
)

type feedback struct {
	url string

	done chan struct{}
	work sync.WaitGroup
	sync.Mutex

	logger.Instance
}

func newFeedback(url string) *feedback {
	f := &feedback{
		url: url,

		Instance: logger.MakeInstance(),
	}

	return f
}

func (f *feedback) Start() {
	f.Lock()
	defer f.Unlock()

	if f.done != nil {
		return
	}
	f.done = make(chan struct{})

	f.work.Add(1)
	go f.background()

	f.Log.Info("started")
}

func (f *feedback) Stop() {
	f.Lock()
	defer f.Unlock()

	if f.done == nil {
		return
	}
	close(f.done)

	f.work.Wait()
	f.done = nil

	f.Log.Info("stopped")
}

func (f *feedback) background() {
	defer f.work.Done()
	var (
		client *ethclient.Client
		err    error
		known  *big.Int
	)

	for {

	connecting:
		for {
			if client != nil {
				client.Close()
				client = nil
			}

			select {
			case <-time.After(time.Second):
			case <-f.done:
				return
			}

			client, err = ethclient.Dial(f.url)
			if err != nil {
				f.Log.Error("connect to", "url", f.url, "err", err)
				client = nil
				continue connecting
			}

			break connecting
		}

	fetching:
		for {

			select {
			case <-time.After(time.Millisecond):
			case <-f.done:
				return
			}

			header, err := client.HeaderByNumber(context.TODO(), nil)
			if err != nil {
				f.Log.Error("HeaderByNumber", "err", err)
				break fetching
			}
			if known == nil {
				known = new(big.Int)
				known.Sub(header.Number, big.NewInt(1))
			}

			for ; header.Number.Cmp(known) > 0; known.Add(known, big.NewInt(1)) {
				// NOTE: err="server returned empty transaction list but block header indicates transactions"
				block, err := client.BlockByNumber(context.TODO(), known)
				if err != nil {
					f.Log.Error("BlockByNumber", "err", err)
					break fetching
				}

				for _, txn := range block.Transactions() {
					info, err := meta.ParseInfo(txn.Data())
					if err != nil || info == nil {
						continue
					}
					f.Log.Info(">>>>>>>>> GOT", "info", info)
				}
			}

		}

	}
}
