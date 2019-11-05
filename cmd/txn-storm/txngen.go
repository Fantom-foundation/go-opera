package main

import (
	"sync"

	"github.com/ethereum/go-ethereum/core/types"

	"github.com/Fantom-foundation/go-lachesis/inter/pos"
	"github.com/Fantom-foundation/go-lachesis/logger"
)

type genThread struct {
	acc      *Acc
	accs     []*Acc
	offset   uint
	position uint

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
			g.generate(g.position)
			g.position++
		}
	}
}

func (g *genThread) generate(position uint) {
	total := uint(len(g.accs))

	if position < total && g.accs[position] == nil {
		b := MakeAcc(position + g.offset)
		nonce := position + 1
		amount := pos.StakeToBalance(100)
		txn := g.acc.TransactionTo(b, nonce, amount)
		g.send(txn)
		g.accs[position] = b

		g.Log.Info("initial txn", "from", "donor", "to", position+g.offset)
		return
	}

	a := position % total
	b := (position + 1) % total
	nonce := position/total + 1
	amount := pos.StakeToBalance(10)
	txn := g.accs[a].TransactionTo(g.accs[b], nonce, amount)
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
