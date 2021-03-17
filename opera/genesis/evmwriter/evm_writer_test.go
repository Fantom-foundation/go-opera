package evmwriter

import (
	"bytes"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/stretchr/testify/require"
)

func TestSign(t *testing.T) {
	require := require.New(t)

	parsed, err := abi.JSON(strings.NewReader(ContractABI))
	require.NoError(err)

	for name, sign := range map[string][]byte{
		"setBalance": []byte{0xe3, 0x04, 0x43, 0xbc},
		"copyCode":   []byte{0xd6, 0xa0, 0xc7, 0xaf},
		"swapCode":   []byte{0x07, 0x69, 0x0b, 0x2a},
		"setStorage": []byte{0x39, 0xe5, 0x03, 0xab},
	} {
		method, exist := parsed.Methods[name]
		require.True(exist)
		require.True(
			bytes.Equal(method.ID, sign),
		)
	}

}
