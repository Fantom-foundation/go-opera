package main

import (
	"context"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/Fantom-foundation/go-lachesis/cmd/tx-storm/meta"
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

func (f *feedback) Start() <-chan big.Int {
	f.Lock()
	defer f.Unlock()

	if f.done != nil {
		return nil
	}
	f.done = make(chan struct{})

	blocks := make(chan big.Int, 1)
	f.work.Add(1)
	go f.background(blocks)

	f.Log.Info("started")
	return blocks
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

func (f *feedback) background(blocks chan<- big.Int) {
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
			case <-time.After(time.Second):
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

			f.Log.Info("header", "num", header.Number)

			for ; header.Number.Cmp(known) > 0; known.Add(known, big.NewInt(1)) {
				block, err := client.BlockByNumber(context.TODO(), known)
				if err != nil {
					f.Log.Error("BlockByNumber", "err", err)
					break fetching
				}

				select {
				case blocks <- *known:
					f.Log.Info("block", "num", header.Number, "txs", block.Transactions())
				case <-f.done:
					return
				}

				for _, tx := range block.Transactions() {
					info, err := meta.ParseInfo(tx.Data())
					if err != nil {
						f.Log.Error("meta.ParseInfo", "err", err)
						continue
					}
					if info == nil {
						f.Log.Info("3rd-party tx", "tx", tx)
						continue
					}

					f.Log.Info(">>>>>>>>> GOT", "info", info)
				}
			}

		}

	}
}
