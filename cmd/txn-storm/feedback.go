package main

import (
	"context"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

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
		client       *ethclient.Client
		err          error
		newHead      = make(chan *types.Header)
		subscription ethereum.Subscription
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

			subscription, err = client.SubscribeNewHead(context.TODO(), newHead)
			if err != nil {
				f.Log.Error("SubscribeNewHead", "err", err)
				continue connecting
			}

			break connecting
		}

	listening:
		for {
			select {
			case err := <-subscription.Err():
				f.Log.Error("NewHeadSubscription", "err", err)
				break listening
			case header := <-newHead:
				f.Log.Info(">>>>>>>> NewHead", "header", header)
			case <-f.done:
				return
			}
		}

	}
}

/*
func (f *feedback) readBlock(ctx context.Context) {
	header, err := f.client.HeaderByNumber(ctx, nil)
	if err != nil {
		f.Log.Error("HeaderByNumber", "err", err)
		return
	}

	f.client.SubscribeNewHead()

	header.Number
}
*/
