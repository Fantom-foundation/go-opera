package main

import (
	"context"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/Fantom-foundation/go-lachesis/logger"
)

type feedback struct {
	url     string
	OnGetTx func(types.Transactions)

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

	f.Log.Info("Started")
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

	f.Log.Info("Stopped")
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
				f.Log.Error("Connect to", "url", f.url, "err", err)
				client = nil
				continue connecting
			}

			break connecting
		}

	fetching:
		for {
			header, err := client.HeaderByNumber(context.TODO(), nil)
			if err != nil {
				f.Log.Error("HeaderByNumber", "err", err)
				break fetching
			}
			if known == nil {
				if header.Number.Cmp(big.NewInt(1)) > 0 {
					known = new(big.Int)
					known.Sub(header.Number, big.NewInt(1))
				} else {
					known = big.NewInt(1)
				}
			}

			if header.Number.Cmp(known) <= 0 {
				select {
				case <-time.After(10 * time.Millisecond):
					continue fetching
				case <-f.done:
					return
				}
			}

			for ; header.Number.Cmp(known) > 0; known.Add(known, big.NewInt(1)) {
				f.Log.Info("Header", "num", header.Number)
				block, err := client.BlockByNumber(context.TODO(), known)
				if err != nil {
					f.Log.Error("BlockByNumber", "err", err)
					break fetching
				}

				select {
				case blocks <- *known:
					f.Log.Info("Block", "num", header.Number, "txs", block.Transactions().Len())
				case <-f.done:
					return
				}
				f.OnGetTx(block.Transactions())
			}

		}

	}
}
