package main

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type threads struct {
	url string
	all []*genThread

	input  chan *types.Transaction
	output chan struct{}
	work   sync.WaitGroup
	sync.Mutex
}

func newThreads(
	nodeUrl string,
	donor uint,
	num, ofTotal uint,
	maxTxnsPerSec uint,
	block time.Duration,
) *threads {
	if num < 1 || num > ofTotal {
		panic("num is a generator number of total generators count")
	}

	count := runtime.NumCPU()
	runtime.GOMAXPROCS(count)

	tt := &threads{
		url: nodeUrl,
		all: make([]*genThread, count, count),
	}

	accs := maxTxnsPerSec * uint(block.Milliseconds()/1000) / ofTotal
	accsOnThread := accs / uint(count)

	from := accs * (num)
	for i := range tt.all {
		tt.all[i] = newTxnGenerator(donor, from, from+accsOnThread)
		from += accsOnThread
	}

	return tt
}

func (tt *threads) Start() {
	tt.Lock()
	defer tt.Unlock()

	if tt.input != nil {
		return
	}

	tt.input = make(chan *types.Transaction, 100)

	for _, t := range tt.all {
		t.SetOutput(tt.input)
		t.Start()
	}

	tt.output = make(chan struct{})
	tt.work.Add(1)
	go tt.background()
}

func (tt *threads) Stop() {
	tt.Lock()
	defer tt.Unlock()

	if tt.input == nil {
		return
	}

	var stoped sync.WaitGroup
	stoped.Add(len(tt.all))
	for _, t := range tt.all {
		go func(t *genThread) {
			t.Stop()
			stoped.Done()
		}(t)
	}
	stoped.Wait()

	close(tt.input)
	close(tt.output)
	tt.work.Wait()
	tt.input = nil
}

func (tt *threads) background() {
	defer tt.work.Done()

	client, err := ethclient.Dial(tt.url)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	for txn := range tt.input {
		for {
			err := client.SendTransaction(context.Background(), txn)
			if err == nil {
				break
			}

			fmt.Println(err)
			select {
			case <-time.After(2 * time.Second):
				continue
			case <-tt.output:
				return
			}
		}
		fmt.Println(txn)
	}
}
