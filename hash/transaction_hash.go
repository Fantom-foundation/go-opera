package hash

import (
	"math/rand"

	"github.com/ethereum/go-ethereum/common"
)

type (
	// Transaction is a unique identifier of internal transaction.
	// It is a hash of Transaction.
	Transaction common.Hash
)

var (
	// ZeroTransaction is a hash of virtual initial transaction.
	ZeroTransaction = Transaction{}
)

// HexToTransactionHash sets byte representation of s to hash.
// If b is larger than len(h), b will be cropped from the left.
func HexToTransactionHash(s string) Transaction {
	return Transaction(common.HexToHash(s))
}

// Bytes returns value as byte slice.
func (h Transaction) Bytes() []byte {
	return (common.Hash)(h).Bytes()
}

// Hex converts an event hash to a hex string.
func (h Transaction) Hex() string {
	return common.Hash(h).Hex()
}

// IsZero returns true if hash is empty.
func (h *Transaction) IsZero() bool {
	return *h == Transaction{}
}

/*
 * Utils:
 */

// FakeTransaction generates random fake hash for testing purpose.
func FakeTransaction() Transaction {
	return Transaction(FakeHash())
}

// FakeHash generates random fake hash for testing purpose.
func FakeHash(seed ...int64) (h common.Hash) {
	randRead := rand.Read

	if len(seed) > 0 {
		src := rand.NewSource(seed[0])
		rnd := rand.New(src)
		randRead = rnd.Read
	}

	_, err := randRead(h[:])
	if err != nil {
		panic(err)
	}
	return
}
