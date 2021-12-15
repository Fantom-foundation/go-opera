package evmwriter

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSign(t *testing.T) {
	require := require.New(t)

	require.Equal([]byte{0xe3, 0x04, 0x43, 0xbc}, setBalanceMethodID)
	require.Equal([]byte{0xd6, 0xa0, 0xc7, 0xaf}, copyCodeMethodID)
	require.Equal([]byte{0x07, 0x69, 0x0b, 0x2a}, swapCodeMethodID)
	require.Equal([]byte{0x39, 0xe5, 0x03, 0xab}, setStorageMethodID)
	require.Equal([]byte{0x79, 0xbe, 0xad, 0x38}, incNonceMethodID)

}
