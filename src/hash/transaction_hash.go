package hash

type (
	// Transaction is a unique identifier of internal transaction.
	// It is a hash of Transaction.
	Transaction Hash
)

var (
	// ZeroTransaction is a hash of virtual initial transaction.
	ZeroTransaction = Transaction{}
)

// HexToTransactionHash sets byte representation of s to hash.
// If b is larger than len(h), b will be cropped from the left.
func HexToTransactionHash(s string) Transaction {
	return Transaction(HexToHash(s))
}

// Bytes returns value as byte slice.
func (h Transaction) Bytes() []byte {
	return (Hash)(h).Bytes()
}

// Hex converts an event hash to a hex string.
func (h Transaction) Hex() string {
	return Hash(h).Hex()
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
