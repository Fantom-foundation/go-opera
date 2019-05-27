package hash

type (
	// InternalTransaction is a unique identifier of internal transaction.
	// It is a hash of InternalTransaction.
	InternalTransaction Hash
)

var (
	// ZeroInternalTransaction is a hash of virtual initial transaction.
	ZeroInternalTransaction = InternalTransaction{}
)

// HexToInternalTransactionHash sets byte representation of s to hash.
// If b is larger than len(h), b will be cropped from the left.
func HexToInternalTransactionHash(s string) InternalTransaction {
	return InternalTransaction(HexToHash(s))
}

// Bytes returns value as byte slice.
func (h InternalTransaction) Bytes() []byte {
	return (Hash)(h).Bytes()
}

// Hex converts an event hash to a hex string.
func (h InternalTransaction) Hex() string {
	return Hash(h).Hex()
}

// IsZero returns true if hash is empty.
func (h *InternalTransaction) IsZero() bool {
	return *h == InternalTransaction{}
}
