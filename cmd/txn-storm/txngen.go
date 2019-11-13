package main

import (
	"math/big"
	"strconv"
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
	const (
		X    = 3
		dose = 10
	)
	var (
		total = uint(len(g.accs))

		txKind   string
		a, b     uint
		aa       string
		from, to *Acc
		nonce    uint
		amount   *big.Int
	)

	if position < total && g.accs[position] == nil {
		txKind = "initial txn"
		b = position
		to = MakeAcc(b + g.offset)
		g.accs[b] = to

		if b == 0 {
			nonce = init
			aa = "donor"
			from = g.donorAcc
		} else {
			nonce = (b-1)%X + 1
			a = (b - 1) / X
			aa = strconv.FormatUint(uint64(a+g.offset), 10)
			from = g.accs[a]
		}

		times := pow(X, (total-b)/X)
		amount = pos.StakeToBalance(pos.Stake(times * dose))

	} else {
		txKind = "regular txn"
		a = position % total
		b = (position + 1) % total
		aa = strconv.FormatUint(uint64(a+g.offset), 10)
		from = g.accs[a]
		to = g.accs[b]
		nonce = 0 // position/total + 2
		amount = pos.StakeToBalance(pos.Stake(dose))
	}

	g.Log.Info(txKind, "nonce", nonce, "from", aa, "to", b+g.offset, "amount", amount)
	txn = from.TransactionTo(to, nonce, amount, meta.NewInfo(aa, b+g.offset).Bytes())
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

func pow(x, y uint) uint {
	res := uint(1)
	for i := uint(0); i < y; i++ {
		res *= x
	}
	return res
}
