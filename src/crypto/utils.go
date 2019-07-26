package crypto

import (
	"fmt"
	"math/big"
	"strings"
)

// RawEncodeSignature encodes signature to string.
// NOTE: deprecated.
func RawEncodeSignature(r, s *big.Int) string {
	return fmt.Sprintf("%s|%s", r.Text(36), s.Text(36))
}

// RawDecodeSignature decodes signature from string.
// NOTE: deprecated.
func RawDecodeSignature(sign string) (r, s *big.Int, err error) {
	values := strings.Split(sign, "|")
	if len(values) != 2 {
		return r, s, fmt.Errorf("wrong number of values in signature: got %d, want 2", len(values))
	}
	r, _ = new(big.Int).SetString(values[0], 36)
	s, _ = new(big.Int).SetString(values[1], 36)
	return r, s, nil
}
