package main

import (
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/core/types"

	"github.com/Fantom-foundation/go-lachesis/cmd/tx-storm/meta"
	"github.com/Fantom-foundation/go-lachesis/logger"
)

type Transaction struct {
	Raw  *types.Transaction
	Info *meta.Info
}

type generator struct {
	accs   []*Acc
	offset uint

	position uint

	output chan<- *Transaction

	work sync.WaitGroup
	done chan struct{}
	sync.Mutex

	logger.Instance
}

func newTxGenerator(num, accs, offset uint) *generator {
	g := &generator{
		accs:   make([]*Acc, accs),
		offset: offset,

		Instance: logger.MakeInstance(),
	}

	g.Log.Info("Will use", "accounts", accs, "from", g.offset, "to", accs+g.offset)
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
			tx := g.Yield()
			g.send(tx)
		}
	}
}

func (g *generator) Yield() *Transaction {
	tx := g.generate(g.position)
	g.position++

	return tx
}

func (g *generator) generate(position uint) *Transaction {
	var count = uint(len(g.accs))

	a := position % count
	b := (position + 1) % count

	from := g.accs[a]
	if from == nil {
		from = MakeAcc(a + g.offset)
		g.accs[a] = from
	}
	a += g.offset

	to := g.accs[b]
	if to == nil {
		to = MakeAcc(b + g.offset)
		g.accs[b] = to
	}
	b += g.offset

	nonce := position / count
	amount := big.NewInt(1e6)

	tx := &Transaction{
		Raw:  from.TransactionTo(to, nonce, amount),
		Info: meta.NewInfo(a, b),
	}

	g.Log.Info("Regular tx", "from", a, "to", b, "amount", amount, "nonce", nonce)
	return tx
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
