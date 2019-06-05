package hash

import (
	"crypto/sha256"
	"fmt"
	"math/big"
	"math/rand"
	"reflect"

	"golang.org/x/crypto/sha3"

	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/common/hexutil"
)

const (
	// Size is the expected length of the hash
	Size = sha256.Size
)

var (
	hashT = reflect.TypeOf(Hash{})
)

// Hash represents the 32 byte Keccak256 hash of arbitrary data.
type Hash [Size]byte

func Of(data ...[]byte) (hash Hash) {
	d := sha3.NewLegacyKeccak256()
	for _, b := range data {
		_, err := d.Write(b)
		if err != nil {
			panic(err)
		}
	}
	d.Sum(hash[:0])
	return hash
}

// FromBytes converts bytes to hash.
// If b is larger than len(h), b will be cropped from the left.
func FromBytes(b []byte) Hash {
	var h Hash
	h.SetBytes(b)
	return h
}

// BigToHash sets byte representation of b to hash.
// If b is larger than len(h), b will be cropped from the left.
func BigToHash(b *big.Int) Hash { return FromBytes(b.Bytes()) }

// HexToHash sets byte representation of s to hash.
// If b is larger than len(h), b will be cropped from the left.
func HexToHash(s string) Hash { return FromBytes(common.FromHex(s)) }

// Bytes gets the byte representation of the underlying hash.
func (h Hash) Bytes() []byte { return h[:] }

// Big converts a hash to a big integer.
func (h Hash) Big() *big.Int { return new(big.Int).SetBytes(h[:]) }

// Hex converts a hash to a hex string.
func (h Hash) Hex() string { return hexutil.Encode(h[:]) }

// TerminalString implements log.TerminalStringer, formatting a string for console
// output during logging.
func (h Hash) TerminalString() string {
	return fmt.Sprintf("%x…%x", h[:3], h[29:])
}

// String implements the stringer interface and is used also by the logger when
// doing full logging into a file.
func (h Hash) String() string {
	return h.Hex()
}

// ShortString returns short string representation.
func (h Hash) ShortString() string {
	return hexutil.Encode(h[:3]) + "..."
}

// Format implements fmt.Formatter, forcing the byte slice to be formatted as is,
// without going through the stringer interface used for logging.
func (h Hash) Format(s fmt.State, c rune) {
	if _, err := fmt.Fprintf(s, "%"+string(c), h[:]); err != nil {
		panic(err)
	}
}

// UnmarshalText parses a hash in hex syntax.
func (h *Hash) UnmarshalText(input []byte) error {
	return hexutil.UnmarshalFixedText("Hash", input, h[:])
}

// UnmarshalJSON parses a hash in hex syntax.
func (h *Hash) UnmarshalJSON(input []byte) error {
	return hexutil.UnmarshalFixedJSON(hashT, input, h[:])
}

// MarshalText returns the hex representation of h.
func (h Hash) MarshalText() ([]byte, error) {
	return hexutil.Bytes(h[:]).MarshalText()
}

// SetBytes converts bytes to hash.
// If b is larger than len(h), b will be cropped from the left.
func (h *Hash) SetBytes(raw []byte) {
	copy(h[:], raw)
	for i := len(raw); i < len(h); i++ {
		h[i] = 0
	}
}

// Generate implements testing/quick.Generator.
func (h Hash) Generate(rand *rand.Rand, size int) reflect.Value {
	m := rand.Intn(len(h))
	for i := len(h) - 1; i > m; i-- {
		h[i] = byte(rand.Uint32())
	}
	return reflect.ValueOf(h)
}

// Scan implements Scanner for database/sql.
func (h *Hash) Scan(src interface{}) error {
	srcB, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("can't scan %T into Hash", src)
	}
	if len(srcB) != Size {
		return fmt.Errorf("can't scan []byte of len %d into Hash, want %d", len(srcB), Size)
	}
	copy(h[:], srcB)
	return nil
}

/*
 * Utils:
 */

// FakeHash generates random fake hash for testing purpose.
func FakeHash() (h Hash) {
	_, err := rand.Read(h[:])
	if err != nil {
		panic(err)
	}
	return
}
