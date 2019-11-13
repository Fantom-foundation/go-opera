package main

import (
	"fmt"
	"math/big"
	"runtime"
	"sync"
	"time"

	"github.com/Fantom-foundation/go-lachesis/logger"
)

type threads struct {
	generators []*generator
	senders    []*sender
	feedback   *feedback

	maxTxnsPerSec uint

	done chan struct{}
	work sync.WaitGroup
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
		generators: make([]*generator, count, count),
		senders:    make([]*sender, 10, 10),

		Instance: logger.MakeInstance(),
	}
	tt.SetName("Threads")

	tt.maxTxnsPerSec = maxTxnsPerSec / ofTotal
	accs := tt.maxTxnsPerSec * uint(block.Milliseconds()/1000)
	accsOnThread := accs / uint(count)

	from := accs * num
	for i := range tt.generators {
		tt.generators[i] = newTxGenerator(donor, from, from+accsOnThread)
		tt.generators[i].SetName(fmt.Sprintf("Generator-%d-%d", from, from+accsOnThread))
		from += accsOnThread
	}

	for i := range tt.senders {
		tt.senders[i] = newSender(nodeUrl)
		tt.senders[i].SetName(fmt.Sprintf("Sender%d", i))
	}

	tt.feedback = newFeedback(nodeUrl)
	tt.feedback.SetName("Feedback")

	return tt
}

func (tt *threads) Start() {
	tt.Lock()
	defer tt.Unlock()

	if tt.done != nil {
		return
	}

	destinations := make([]chan<- *Transaction, len(tt.senders))
	for i, s := range tt.senders {
		destination := make(chan *Transaction, X)
		destinations[i] = destination
		s.Start(destination)
	}
	source := make(chan *Transaction, len(tt.generators)*2)
	tt.done = make(chan struct{})
	tt.work.Add(1)
	go tt.txTransfer(source, destinations)

	for i, t := range tt.generators {
		// first transactions from donor: one after another
		tx := t.Yield(uint(i))
		destinations[0] <- tx
	}
	for _, t := range tt.generators {
		t.Start(source)
	}

	blocks := tt.feedback.Start()
	tt.work.Add(1)
	go tt.blockNotify(blocks, tt.senders)

	tt.Log.Info("started")
}

func (tt *threads) Stop() {
	tt.Lock()
	defer tt.Unlock()

	if tt.done == nil {
		return
	}

	var stoped sync.WaitGroup
	stoped.Add(1)
	go func() {
		tt.feedback.Stop()
		stoped.Done()
	}()
	stoped.Add(len(tt.generators))
	for _, t := range tt.generators {
		go func(t *generator) {
			t.Stop()
			stoped.Done()
		}(t)
	}
	stoped.Add(len(tt.senders))
	for _, s := range tt.senders {
		go func(s *sender) {
			s.Stop()
			stoped.Done()
		}(s)
	}
	stoped.Wait()

	close(tt.done)
	tt.work.Wait()
	tt.done = nil

	tt.Log.Info("stopped")
}

func (tt *threads) blockNotify(blocks <-chan big.Int, senders []*sender) {
	defer tt.work.Done()
	for {
		select {
		case bnum := <-blocks:
			for _, s := range senders {
				s.Notify(bnum)
			}
		case <-tt.done:
			return
		}
	}
}

func (tt *threads) txTransfer(
	source <-chan *Transaction,
	destinations []chan<- *Transaction,
) {
	defer tt.work.Done()
	defer func() {
		for _, d := range destinations {
			close(d)
		}
	}()

	var (
		count uint
		start time.Time
		tx    *Transaction
	)
	for {

		if time.Since(start) >= time.Second {
			count = 0
			start = time.Now()
		}

		if count >= tt.maxTxnsPerSec {
			timeout := start.Add(time.Second).Sub(time.Now())
			tt.Log.Debug("tps limit", "timeout", timeout)
			select {
			case <-time.After(timeout):
			case <-tt.done:
				return
			}
		}

		select {
		case tx = <-source:
			count++
		case <-tt.done:
			return
		}

		// the same from-addr to the same sender
		destination := destinations[tx.Info.From%uint(len(destinations))]
		select {
		case destination <- tx:
		case <-tt.done:
			return
		}

	}
}
