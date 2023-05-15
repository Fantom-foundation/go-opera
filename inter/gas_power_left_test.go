package inter

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/stretchr/testify/require"
)

func TestRlp(t *testing.T) {
	require := require.New(t)
	v := GasPowerLeft{
		Gas: [2]uint64{0xBAADCAFE, 0xDEADBEEF},
	}
	b, err := rlp.EncodeToBytes(v)
	require.NoError(err)
	require.Equal("cbca84baadcafe84deadbeef", common.Bytes2Hex(b))
}
