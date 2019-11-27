package main

import (
	"fmt"
	"math/big"
	"runtime"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/core/types"

	"github.com/Fantom-foundation/go-lachesis/logger"
)

type threads struct {
	generators []*generator
	senders    []*sender
	feedback   *feedback
	txs        *txsDict

	maxTxnsPerSec uint

	done chan struct{}
	work sync.WaitGroup
	sync.Mutex

	logger.Instance
}

func newThreads(
	nodeUrl string,
	num, ofTotal uint,
	maxTxnsPerSec uint,
	period time.Duration,
) *threads {
	if num >= ofTotal {
		panic("num is a generator number of total generators count")
	}

	count := runtime.NumCPU()
	runtime.GOMAXPROCS(count)

	tt := &threads{
		generators: make([]*generator, count, count),
		senders:    make([]*sender, 10, 10),
		txs:        newTxsDict(),

		Instance: logger.MakeInstance(),
	}

	tt.maxTxnsPerSec = maxTxnsPerSec / ofTotal
	accs := tt.maxTxnsPerSec * uint(period.Milliseconds()/1000)
	accsOnThread := approximate(accs / uint(count))
	accs = accsOnThread * uint(count)
	offset := accs * (num + 1)
	tt.Log.Info("Will create", "accounts", accs, "from", offset, "to", accs+offset)

	donor := num
	for i := range tt.generators {
		tt.generators[i] = newTxGenerator(donor, uint(i), accsOnThread, offset)
		tt.generators[i].SetName(fmt.Sprintf("Generator-%d", i))
	}

	for i := range tt.senders {
		tt.senders[i] = newSender(nodeUrl)
		tt.senders[i].OnSendTx = tt.onSendTx
		tt.senders[i].SetName(fmt.Sprintf("Sender%d", i))
	}

	tt.feedback = newFeedback(nodeUrl)
	tt.feedback.OnGetTx = tt.onGetTx
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
		destination := make(chan *Transaction, multiplicator)
		destinations[i] = destination
		s.Start(destination)
	}
	source := make(chan *Transaction, len(tt.generators)*2)
	tt.done = make(chan struct{})
	tt.work.Add(1)
	go tt.txTransfer(source, destinations)

	for i, t := range tt.generators {
		// first transactions from donor: one by one
		tx := t.Yield(uint(i))
		destinations[0] <- tx
	}

	for _, t := range tt.generators {
		t.Start(source)
	}

	blocks := tt.feedback.Start()
	tt.work.Add(1)
	go tt.blockNotify(blocks, tt.senders)

	tt.Log.Info("Started")
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

	tt.Log.Info("Stopped")
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

func (tt *threads) onSendTx(tx *Transaction) {
	txCountSentMeter.Inc(1)
	tt.txs.Start(tx.Raw.Hash())
}

func (tt *threads) onGetTx(txs types.Transactions) {
	txCountGotMeter.Inc(int64(txs.Len()))

	for _, tx := range txs {
		latency, err := tt.txs.Finish(tx.Hash())
		if err != nil {
			continue
		}
		txLatencyMeter.Update(latency.Milliseconds())
	}
}
