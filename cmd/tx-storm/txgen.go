package main

import (
	"math/big"
	"sync"

	"github.com/Fantom-foundation/go-lachesis/cmd/tx-storm/meta"
	"github.com/Fantom-foundation/go-lachesis/inter/pos"
	"github.com/Fantom-foundation/go-lachesis/logger"
)

type generator struct {
	donor  *Acc
	accs   []*Acc
	offset uint

	position uint

	level struct {
		Num        uint
		Counter    uint
		CurrCount  uint
		NextsCount uint
	}

	output chan<- *Transaction

	work sync.WaitGroup
	done chan struct{}
	sync.Mutex

	logger.Instance
}

func newTxGenerator(donor, num, accs, offset uint) *generator {
	accs = approximate(accs)

	g := &generator{
		donor:  MakeAcc(donor),
		accs:   make([]*Acc, accs),
		offset: offset + accs*num,

		Instance: logger.MakeInstance(),
	}

	g.level.NextsCount = accs

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
			tx, _ := g.Yield(0)
			g.send(tx)
		}
	}
}

func (g *generator) Yield(init uint) (*Transaction, *meta.Info) {
	if g.level.Counter == 0 {
		g.level.CurrCount = pow(multiplicator, g.level.Num)
		g.level.NextsCount -= g.level.CurrCount
		g.level.Counter = g.level.CurrCount
		g.level.Num++
	}
	g.level.Counter--

	tx, info := g.generate(init, g.position)
	g.position++

	return tx, info
}

const multiplicator = 3

func (g *generator) generate(init, position uint) (*Transaction, *meta.Info) {
	const reserve = 10

	var (
		count = uint(len(g.accs))

		txKind   string
		a, b     uint
		from, to *Acc
		nonce    uint
		amount   *big.Int
	)

	if position < count && g.accs[position] == nil {
		txKind = "Initial tx"

		b = position
		to = MakeAcc(b + g.offset)
		g.accs[b] = to

		var childs uint
		if b == 0 {
			nonce = init
			a = 0
			from = g.donor
			childs = count
		} else {
			nonce = (b - 1) % multiplicator
			a = (b - 1) / multiplicator
			from = g.accs[a]
			childs = g.level.NextsCount/g.level.CurrCount + 1
			a += g.offset
		}
		amount = pos.StakeToBalance(pos.Stake(childs * reserve))

	} else {
		txKind = "Regular tx"
		a = position % count
		b = (position + 1) % count
		from = g.accs[a]
		a += g.offset
		to = g.accs[b]
		nonce = position/count + multiplicator - 1
		amount = pos.StakeToBalance(pos.Stake(1))
	}

	b += g.offset
	info := meta.NewInfo(a, b)
	g.Log.Info(txKind, "from", a, "to", b, "nonce", nonce, "amount", amount)
	return from.TransactionTo(to, nonce, amount, info), info
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

func approximate(target uint) (count uint) {
	var level uint
	for target > count {
		count += pow(multiplicator, level)
		level++
	}

	return
}

func pow(x, y uint) uint {
	res := uint(1)
	for i := uint(0); i < y; i++ {
		res *= x
	}

	return res
}
