package lachesis

import (
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/network"
	"github.com/Fantom-foundation/go-lachesis/src/poset"
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
		ctrl, _, err := proxy.NewGrpcCtrlProxy(
			l.CtrlListenAddr(),
			l.node,
			l.consensus,
			nil,
			l.service.listen,
		)
		if err != nil {
			l.Fatal(err)
		}
		defer ctrl.Close()

		<-done
	}(l.service.done)

	go func(done chan struct{}) {
		app, _, err := proxy.NewGrpcAppProxy(
			l.AppListenAddr(),
			l.conf.Node.ClientTimeout,
			l.service.listen,
		)
		if err != nil {
			l.Fatal(err)
		}
		defer app.Close()

		err = l.consensus.OnNewBlock(func(blockNumber idx.Block) {
			b := l.consensusStore.GetBlock(blockNumber)
			block := l.toLegacyBlock(b)
			_, _ = app.CommitBlock(*block)
		}, false)
		if err != nil {
			l.Fatal("unable to register new block callback", err)
		}
		defer l.consensus.OnNewBlock(nil, true)

		for {
			select {
			case tx := <-app.SubmitCh():
				l.node.AddExternalTxn(tx)
			case tx := <-app.SubmitInternalCh():
				l.node.AddInternalTxn(tx)
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
func (l *Lachesis) toLegacyBlock(b *inter.Block) *poset.Block {
	var txns [][]byte
	for _, e := range b.Events {
		event := l.nodeStore.GetEvent(e)
		txns = append(txns, event.ExternalTransactions.Value...)
	}
	// NOTE: Signatures and Hashes are empty
	return &poset.Block{
		Body: &poset.BlockBody{
			Index:        int64(b.Index),
			Transactions: txns,
		},
	}
}
