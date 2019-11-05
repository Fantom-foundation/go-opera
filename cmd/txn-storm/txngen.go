package main

import (
	"sync"

	"github.com/ethereum/go-ethereum/core/types"

	"github.com/Fantom-foundation/go-lachesis/logger"
)

type genThread struct {
	acc    *Acc
	accs   []*Acc
	offset uint
	pos    uint

	output chan<- *types.Transaction

	work sync.WaitGroup
	done chan struct{}
	sync.Mutex

	logger.Instance
}

func newTxnGenerator(donor, from, to uint) *genThread {
	if from >= to {
		panic("invalid range from-to")
	}

	if donor >= from && donor < to {
		panic("donor is in range from-to")
	}

	count := to - from

	g := &genThread{
		acc:    MakeAcc(donor),
		accs:   make([]*Acc, count, count),
		offset: from,

		Instance: logger.MakeInstance(),
	}

	return g
}

func (g *genThread) SetOutput(c chan<- *types.Transaction) {
	g.Lock()
	defer g.Unlock()

	if g.output != nil || g.done != nil {
		panic("do it once before start")
	}

	g.output = c
}

func (g *genThread) Start() {
	g.Lock()
	defer g.Unlock()

	if g.done != nil {
		return
	}

	g.done = make(chan struct{})
	g.work.Add(1)
	go g.background()

	g.Log.Info("started")
}

func (g *genThread) Stop() {
	g.Lock()
	defer g.Unlock()

	if g.done == nil {
		return
	}

	close(g.done)
	g.work.Wait()
	g.done = nil

	g.Log.Info("stopped")
}

func (g *genThread) background() {
	defer g.work.Done()
	for {
		select {
		case <-g.done:
			return
		default:
			g.generate(g.pos)
			g.pos++
		}
	}
}

func (g *genThread) generate(pos uint) {
	total := uint(len(g.accs))

	if pos < total && g.accs[pos] == nil {
		b := MakeAcc(pos + g.offset)
		nonce := pos + 1
		txn := g.acc.TransactionTo(b, nonce, 10)
		g.send(txn)
		g.accs[pos] = b

		g.Log.Info("initial txn", "from", "donor", "to", pos+g.offset)
		return
	}

	a := pos % total
	b := (pos + 1) % total
	nonce := pos/total + 1
	txn := g.accs[a].TransactionTo(g.accs[b], nonce, 1)
	g.send(txn)

	g.Log.Info("regular txn", "from", a+g.offset, "to", b+g.offset)
}

func (g *genThread) send(tx *types.Transaction) {
	if g.output == nil {
		return
	}

	select {
	case g.output <- tx:
		break
	case <-g.done:
		break
	}
}
