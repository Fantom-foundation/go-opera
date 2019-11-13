package main

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/Fantom-foundation/go-lachesis/evm_core"
	"github.com/Fantom-foundation/go-lachesis/logger"
)

type sender struct {
	url    string
	input  <-chan *Transaction
	blocks chan big.Int

	done chan struct{}
	work sync.WaitGroup
	sync.Mutex

	logger.Instance
}

func newSender(url string) *sender {
	return &sender{
		url:    url,
		blocks: make(chan big.Int, 1),

		Instance: logger.MakeInstance(),
	}
}

func (s *sender) Start(c <-chan *Transaction) {
	s.Lock()
	defer s.Unlock()

	if s.done != nil {
		return
	}

	s.input = c
	s.done = make(chan struct{})
	s.work.Add(1)
	go s.background()

	s.Log.Info("started")
}

func (s *sender) Stop() {
	s.Lock()
	defer s.Unlock()

	if s.done == nil {
		return
	}

	close(s.done)
	s.work.Wait()
	s.done = nil

	s.Log.Info("stopped")
}

func (s *sender) background() {
	defer s.work.Done()

	var (
		client *ethclient.Client
		err    error
		tx     *Transaction
		info   string
	)

	for {
		select {
		case tx = <-s.input:
			info = tx.Info.String()
		case <-s.done:
			return
		}

	connecting:
		for client == nil {
			client, err = ethclient.Dial(s.url)
			if err != nil {
				client = nil
				s.Log.Error("connect to", "url", s.url, "err", err)
				select {
				case <-time.After(time.Second):
					continue connecting
				case <-s.done:
					return
				}
			}
		}

	sending:
		for {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			err = client.SendTransaction(ctx, tx.Tx)
			cancel()
			if err == nil {
				s.Log.Info("txn sending ok", "info", info, "amount", tx.Tx.Value())
				txnsCountMeter.Inc(1)
				break sending
			}

			switch err.Error() {
			case fmt.Sprintf("known transaction: %x", tx.Tx.Hash()):
				s.Log.Info("txn sending skip", "info", info, "amount", tx.Tx.Value(), "cause", err)
				break sending
			case evm_core.ErrNonceTooLow.Error():
				s.Log.Info("txn sending skip", "info", info, "amount", tx.Tx.Value(), "cause", err)
				break sending
			case evm_core.ErrReplaceUnderpriced.Error():
				s.Log.Info("txn sending skip", "info", info, "amount", tx.Tx.Value(), "cause", err)
				break sending
			default:
				s.Log.Error("txn sending err", "info", info, "amount", tx.Tx.Value(), "cause", err)
				select {
				case <-s.blocks:
					s.Log.Error("try to send txn again", "info", info, "amount", tx.Tx.Value())
				case <-s.done:
					return
				}
			}
		}

	}
}

func (s *sender) Notify(bnum big.Int) {
	select {
	case s.blocks <- bnum:
	default:
	}
}
