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
	url    string
	Ready  chan struct{}
	client *ethclient.Client

	done chan struct{}
	work sync.WaitGroup

	logger.Instance
}

func newSender(url string) *sender {
	s := &sender{
		url:   url,
		Ready: make(chan struct{}, 1),
		done:  make(chan struct{}),

		Instance: logger.MakeInstance(),
	}
	s.Ready <- struct{}{}

	return s
}

func (s *sender) Close() {
	close(s.done)
	s.work.Wait()
	close(s.Ready)
	if s.client != nil {
		s.client.Close()
	}
}

// TODO: prevent double calls
func (s *sender) Send(txn *types.Transaction) {
	s.work.Add(1)
	go func() {
		defer s.work.Done()
		var err error

	connecting:
		for s.client == nil {
			s.client, err = ethclient.Dial(s.url)
			if err == nil {
				break connecting
			}

			s.Log.Error("connect to", "url", s.url, "err", err)
			select {
			case <-time.After(time.Second):
			case <-s.done:
				return
			}
		}

	sending:
		for {
			err = s.client.SendTransaction(context.Background(), txn)
			if err == nil {
				s.Log.Info("txn sending ok")
				txnsCountMeter.Inc(1)
				break sending
			}

			switch err.Error() {
			case fmt.Sprintf("known transaction: %x", txn.Hash()):
				break sending
			case evm_core.ErrNonceTooLow.Error():
				break sending
			case evm_core.ErrReplaceUnderpriced.Error():
				break sending
			default:
				s.Log.Error("txn sending", "err", err, "txn", txnToJson(txn))
			}

			select {
			case <-time.After(time.Second):
			case <-s.done:
				return
			}
		}

		s.Ready <- struct{}{}
	}()
}

func txnToJson(txn *types.Transaction) string {
	b, err := txn.MarshalJSON()
	if err != nil {
		panic(err)
	}
	return string(b)
}
