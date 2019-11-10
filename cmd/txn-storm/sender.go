package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/Fantom-foundation/go-lachesis/evm_core"
	"github.com/Fantom-foundation/go-lachesis/logger"
)

type sender struct {
	url   string
	input <-chan *types.Transaction

	done chan struct{}
	work sync.WaitGroup
	sync.Mutex

	logger.Instance
}

func newSender(url string) *sender {
	s := &sender{
		url: url,

		Instance: logger.MakeInstance(),
	}

	return s
}

func (s *sender) Start(c <-chan *types.Transaction) {
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
	)

	for txn := range s.input {

		//connecting:
		for client == nil {
			client, err = ethclient.Dial(s.url)
			if err != nil {
				client = nil
				s.Log.Error("connect to", "url", s.url, "err", err)
				select {
				case <-time.After(time.Second):
				case <-s.done:
					return
				}
			}
		}

	sending:
		for {
			err = client.SendTransaction(context.Background(), txn)
			if err == nil {
				s.Log.Info("txn sending ok", "data", string(txn.Data()))
				txnsCountMeter.Inc(1)
				break sending
			}

			switch err.Error() {
			case fmt.Sprintf("known transaction: %x", txn.Hash()):
				s.Log.Info("txn sending skip", "data", string(txn.Data()))
				break sending
			case evm_core.ErrNonceTooLow.Error():
				s.Log.Info("txn sending skip", "data", string(txn.Data()))
				break sending
			case evm_core.ErrReplaceUnderpriced.Error():
				s.Log.Info("txn sending skip", "data", string(txn.Data()))
				break sending
			default:
				s.Log.Error("try to send txn again", "cause", err, "txn", string(txn.Data()))
				select {
				case <-time.After(time.Second):
				case <-s.done:
					return
				}
			}
		}

	}
}
