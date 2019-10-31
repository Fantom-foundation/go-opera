package main

import (
	"sync"

	"github.com/ethereum/go-ethereum/core/types"
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
}

func newTxnGenerator(donor, from, to uint) *genThread {
	if from >= to {
		panic("invalid range from-to")
	}

	if donor >= from && donor < to {
		panic("donor is in range from-to")
	}

	count := to - from

	return &genThread{
		acc:    MakeAcc(donor),
		accs:   make([]*Acc, count, count),
		offset: from,
	}
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
		b := MakeAcc(g.offset + pos)
		nonce := pos + 1
		txn := g.acc.TransactionTo(b, nonce, 10)
		g.send(txn)
		g.accs[pos] = b
		return
	}

	a := pos % total
	b := (pos + 1) % total
	nonce := pos/total + 1
	txn := g.accs[a].TransactionTo(g.accs[b], nonce, 1)
	g.send(txn)
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
