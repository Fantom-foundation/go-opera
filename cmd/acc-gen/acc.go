package main

import (
	"os"

	"github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/go-lachesis/crypto"
)

func NewAccs(from, count uint, done <-chan os.Signal) <-chan common.Address {
	accs := make(chan common.Address, 10)

	go func() {
		defer close(accs)

		for i := uint(0); i < count; i++ {
			acc := MakeAcc(i + from)
			select {
			case accs <- acc:
				continue
			case <-done:
				return
			}
		}
	}()

	return accs
}

func MakeAcc(n uint) common.Address {
	key := crypto.FakeKey(int(n))
	addr := crypto.PubkeyToAddress(key.PublicKey)

	return addr
}
