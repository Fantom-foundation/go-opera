package lachesis

import (
	"github.com/Fantom-foundation/go-lachesis/src/network"
	"github.com/Fantom-foundation/go-lachesis/src/poset"
	"github.com/Fantom-foundation/go-lachesis/src/posposet"
	"github.com/Fantom-foundation/go-lachesis/src/proxy"
)

type service struct {
	listen network.ListenFunc
	done   chan struct{}
}

func (l *Lachesis) serviceStart() {
	if l.service.done != nil {
		return
	}
	l.service.done = make(chan struct{})

	go func(done chan struct{}) {
		app, err := proxy.NewGrpcAppProxy(
			l.ListenAddr(),
			l.conf.Node.ClientTimeout,
			nil,
			l.service.listen)
		if err != nil {
			panic(err)
		}
		defer app.Close()

		l.consensus.NewBlockCh = make(chan uint64, 100)

		for {
			select {
			case tx := <-app.SubmitCh():
				l.node.AddExternalTxn(tx)
			case tx := <-app.SubmitInternalCh():
				l.node.AddInternalTxn(tx)
			case num := <-l.consensus.NewBlockCh:
				b := l.consensusStore.GetBlock(num)
				block := l.toLegacyBlock(b)
				app.CommitBlock(*block)
			case <-done:
				return
			}
		}
	}(l.service.done)
}

func (l *Lachesis) serviceStop() {
	if l.service.done == nil {
		return
	}
	close(l.service.done)
	l.service.done = nil
}

// TODO: it is temporary fake solution
func (l *Lachesis) toLegacyBlock(b *posposet.Block) *poset.Block {
	var txns [][]byte
	for _, e := range b.Events {
		event := l.nodeStore.GetEvent(e)
		txns = append(txns, event.ExternalTransactions...)
	}
	// NOTE: Signatures and Hashes are empty
	return &poset.Block{
		Body: &poset.BlockBody{
			Index:        int64(b.Index),
			Transactions: txns,
		},
	}
}
