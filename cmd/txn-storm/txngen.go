package main

import (
	"sync"

	"github.com/ethereum/go-ethereum/core/types"

	"github.com/Fantom-foundation/go-lachesis/cmd/txn-storm/meta"
	"github.com/Fantom-foundation/go-lachesis/inter/pos"
	"github.com/Fantom-foundation/go-lachesis/logger"
)

type generator struct {
	donorNum uint
	donorAcc *Acc
	accs     []*Acc
	offset   uint
	position uint

	output chan<- *types.Transaction

	work sync.WaitGroup
	done chan struct{}
	sync.Mutex

	logger.Instance
}

func newTxnGenerator(donor, from, to uint) *generator {
	if from >= to {
		panic("invalid range from-to")
	}

	if donor >= from && donor < to {
		panic("donor is in range from-to")
	}

	count := to - from

	g := &generator{
		donorNum: donor,
		donorAcc: MakeAcc(donor),
		accs:     make([]*Acc, count, count),
		offset:   from,

		Instance: logger.MakeInstance(),
	}

	return g
}

func (g *generator) Start(c chan<- *types.Transaction) {
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
			txn := g.Yield(1)
			g.send(txn)
		}
	}
}

func (g *generator) Yield(init uint) *types.Transaction {
	txn := g.generate(init, g.position)
	g.position++
	return txn
}

func (g *generator) generate(init, position uint) (txn *types.Transaction) {
	total := uint(len(g.accs))

	if position < total && g.accs[position] == nil {
		b := position
		g.accs[b] = MakeAcc(b + g.offset)

		nonce := init
		a := -1
		from := g.donorAcc
		if b > 0 {
			a = int(b + g.offset - 1)
			from = g.accs[b-1]
		}

		amount := pos.StakeToBalance(pos.Stake((total - b) * 10 * 2))

		txn = from.TransactionTo(g.accs[b], nonce, amount,
			meta.NewInfo(uint(a), b+g.offset).Bytes())

		g.Log.Info("initial txn", "nonce", nonce, "from", a, "to", b+g.offset, "amount", amount)
		return
	}

	a := position % total
	b := (position + 1) % total
	nonce := position/total + 2

	amount := pos.StakeToBalance(10)

	// NOTE: "insufficient funds for gas * price + value" err
	txn = g.accs[a].TransactionTo(g.accs[b], nonce, amount,
		meta.NewInfo(a+g.offset, b+g.offset).Bytes())

	g.Log.Info("regular txn", "nonce", nonce, "from", a+g.offset, "to", b+g.offset, "amount", amount)
	return
}

func (g *generator) send(txn *types.Transaction) {
	if g.output == nil {
		return
	}

	select {
	case g.output <- txn:
		break
	case <-g.done:
		break
	}
}
