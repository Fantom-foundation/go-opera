package main

import (
	"math/big"
	"sync"

	"github.com/Fantom-foundation/go-lachesis/cmd/tx-storm/meta"
	"github.com/Fantom-foundation/go-lachesis/inter/pos"
	"github.com/Fantom-foundation/go-lachesis/logger"
)

type generator struct {
	accs     []*Acc
	offset   uint
	position uint

	output chan<- *Transaction

	work sync.WaitGroup
	done chan struct{}
	sync.Mutex

	logger.Instance
}

func newTxGenerator(from, to uint) *generator {
	if from >= to {
		panic("invalid range from-to")
	}

	count := to - from

	g := &generator{
		accs:   make([]*Acc, count),
		offset: from,

		Instance: logger.MakeInstance(),
	}

	return g
}

func (g *generator) Start(c chan<- *Transaction) {
	g.Lock()
	defer g.Unlock()

	if g.done != nil {
		return
	}

	g.output = c
	g.done = make(chan struct{})
	g.work.Add(1)
	go g.background()

	g.Log.Info("started")
}

func (g *generator) Stop() {
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

func (g *generator) background() {
	defer g.work.Done()

	for {
		select {
		case <-g.done:
			return
		default:
			tx := g.Yield(1)
			g.send(tx)
		}
	}
}

func (g *generator) Yield(init uint) *Transaction {
	tx := g.generate(init, g.position)
	g.position++
	return tx
}

const X = 3

func (g *generator) generate(init, position uint) *Transaction {
	const (
		dose = 10
	)
	var (
		count = uint(len(g.accs))

		txKind   string
		a, b     uint
		from, to *Acc
		nonce    uint
		amount   *big.Int
	)

	txKind = "regular tx"
	a = position % count
	b = (position + 1) % count
	if position < count {
		if g.accs[a] == nil {
			g.accs[a] = MakeAcc(a + g.offset)
		}
		if g.accs[b] == nil {
			g.accs[b] = MakeAcc(b + g.offset)
		}
	}
	from = g.accs[a]
	to = g.accs[b]
	nonce = position / count
	amount = pos.StakeToBalance(pos.Stake(dose))

	a += g.offset
	b += g.offset
	g.Log.Info(txKind, "nonce", nonce, "from", a, "to", b, "amount", amount)
	return from.TransactionTo(to, nonce, amount, meta.NewInfo(a, b))
}

func (g *generator) send(tx *Transaction) {
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
