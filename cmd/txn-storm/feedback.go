package main

import (
	"context"
	"sync"
	"time"

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
		client *ethclient.Client
		err    error
		header *types.Header
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

	listening:
		for {

			select {
			case <-time.After(time.Millisecond):
			case <-f.done:
				return
			}

			newHeader, err := client.HeaderByNumber(context.TODO(), nil)
			if err != nil {
				f.Log.Error("HeaderByNumber", "err", err)
				break listening
			}

			if header != nil && header.Number.Cmp(newHeader.Number) >= 0 {
				continue listening
			}
			header = newHeader

			f.Log.Info(">>>>>>>> NewHead", "header", header)
		}

	}
}
