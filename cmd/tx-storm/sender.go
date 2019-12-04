package main

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/Fantom-foundation/go-lachesis/evmcore"
	"github.com/Fantom-foundation/go-lachesis/logger"
)

type sender struct {
	url      string
	input    <-chan *Transaction
	blocks   chan big.Int
	OnSendTx func(*Transaction)

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

	s.Log.Info("Started")
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

	s.Log.Info("Stopped")
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
				s.Log.Error("Connect to", "url", s.url, "err", err)
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
			err = client.SendTransaction(ctx, tx.Raw)
			cancel()
			if err == nil {
				s.Log.Info("Tx sending ok", "info", info, "amount", tx.Raw.Value(), "nonce", tx.Raw.Nonce())
				s.OnSendTx(tx)
				break sending
			}

			switch err.Error() {
			case fmt.Sprintf("known transaction: %x", tx.Raw.Hash()):
				s.Log.Info("Tx sending skip", "info", info, "amount", tx.Raw.Value(), "cause", err, "nonce", tx.Raw.Nonce())
				break sending
			case evmcore.ErrNonceTooLow.Error():
				s.Log.Info("Tx sending skip", "info", info, "amount", tx.Raw.Value(), "cause", err, "nonce", tx.Raw.Nonce())
				break sending
			case evmcore.ErrReplaceUnderpriced.Error():
				s.Log.Info("Tx sending skip", "info", info, "amount", tx.Raw.Value(), "cause", err, "nonce", tx.Raw.Nonce())
				break sending
			default:
				s.Log.Error("Tx sending err", "info", info, "amount", tx.Raw.Value(), "cause", err, "nonce", tx.Raw.Nonce())
				select {
				case <-s.blocks:
					s.Log.Error("Try to send tx again", "info", info, "amount", tx.Raw.Value(), "nonce", tx.Raw.Nonce())
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
