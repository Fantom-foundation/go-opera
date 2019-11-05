package main

import (
	"context"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

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
		for s.client == nil {
			s.client, err = ethclient.Dial(s.url)
			if err == nil {
				break
			}

			s.Log.Error("connect to", "url", s.url, "err", err)
			select {
			case <-time.After(time.Second):
			case <-s.done:
				return
			}
		}

		for {
			s.Log.Info("try to send", "txn", txnToJson(txn))
			err = s.client.SendTransaction(context.Background(), txn)
			if err == nil {
				s.Log.Info("txn sending ok")
				break
			}

			s.Log.Error("txn sending", "err", err)
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
