package main

import (
	"fmt"
	"reflect"
	"runtime"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/core/types"

	"github.com/Fantom-foundation/go-lachesis/logger"
)

type threads struct {
	url        string
	generators []*genThread

	txnsPerSec uint

	input chan *types.Transaction
	done  chan struct{}
	work  sync.WaitGroup
	sync.Mutex

	logger.Instance
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
		url:        nodeUrl,
		generators: make([]*genThread, count, count),

		Instance: logger.MakeInstance(),
	}

	tt.SetName("Threads")

	tt.txnsPerSec = maxTxnsPerSec / ofTotal
	accs := tt.txnsPerSec * uint(block.Milliseconds()/1000)
	accsOnThread := accs / uint(count)

	from := accs * num
	for i := range tt.generators {
		tt.generators[i] = newTxnGenerator(donor, from, from+accsOnThread)
		tt.generators[i].SetName(fmt.Sprintf("Gen[%d:%d]", from, from+accsOnThread))
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

	tt.input = make(chan *types.Transaction, len(tt.generators)*2)

	for _, t := range tt.generators {
		t.SetOutput(tt.input)
		t.Start()
	}

	tt.done = make(chan struct{})
	tt.work.Add(1)
	go tt.background()

	tt.Log.Info("started")
}

func (tt *threads) Stop() {
	tt.Lock()
	defer tt.Unlock()

	if tt.input == nil {
		return
	}

	var stoped sync.WaitGroup
	stoped.Add(len(tt.generators))
	for _, t := range tt.generators {
		go func(t *genThread) {
			t.Stop()
			stoped.Done()
		}(t)
	}
	stoped.Wait()

	close(tt.input)
	close(tt.done)
	tt.work.Wait()
	tt.input = nil

	tt.Log.Info("stopped")
}

func (tt *threads) background() {
	defer tt.work.Done()

	senders, selectors := tt.makeSenders(10)
	defer tt.closeSenders(senders)
	selectors = append(selectors,
		reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(tt.done),
		})

	count := uint(0)
	start := time.Now()

	for txn := range tt.input {

		if count >= tt.txnsPerSec {
			timeout := start.Add(time.Second).Sub(time.Now())
			tt.Log.Info("sending pause", "timeout", timeout)
			select {
			case <-time.After(timeout):
			case <-tt.done:
				return
			}
		}
		if time.Since(start) >= time.Second {
			count = 0
			start = time.Now()
		}
		count++

		choosen, _, _ := reflect.Select(selectors)
		if choosen < len(senders) {
			senders[choosen].Send(txn)
		} else {
			return
		}

	}
}

func (tt *threads) makeSenders(count int) (
	senders []*sender,
	selectors []reflect.SelectCase,
) {
	senders = make([]*sender, count, count)
	selectors = make([]reflect.SelectCase, count)
	for i := range senders {
		senders[i] = newSender(tt.url)
		senders[i].SetName(fmt.Sprintf("Sender%d", i))
		selectors[i] = reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(senders[i].Ready),
		}
	}
	return
}

func (tt *threads) closeSenders(ss []*sender) {
	for _, s := range ss {
		s.Close()
	}
}
