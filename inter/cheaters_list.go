package inter

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
)

// Cheaters is a slice type for storing cheaters list.
type Cheaters []common.Address

// Len returns the length of s.
func (s Cheaters) Len() int { return len(s) }

// Swap swaps the i'th and the j'th element in s.
func (s Cheaters) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// GetRlp implements Rlpable and returns the i'th element of s in rlp.
func (s Cheaters) GetRlp(i int) []byte {
	enc, _ := rlp.EncodeToBytes(s[i])
	return enc
}
