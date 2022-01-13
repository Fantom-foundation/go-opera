package internaltx

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/status-im/keycard-go/hexutils"
	"github.com/stretchr/testify/require"
)

func TestIsInternal(t *testing.T) {
	require.True(t, IsInternal(types.NewTx(&types.LegacyTx{
		V: new(big.Int),
		R: new(big.Int),
		S: new(big.Int),
	})))
	require.True(t, IsInternal(types.NewTx(&types.DynamicFeeTx{
		V: new(big.Int),
		R: new(big.Int),
		S: new(big.Int),
	})))
	require.True(t, IsInternal(types.NewTx(&types.LegacyTx{
		V: new(big.Int),
		R: new(big.Int),
		S: big.NewInt(1),
	})))
	require.True(t, IsInternal(types.NewTx(&types.LegacyTx{
		V: new(big.Int),
		R: new(big.Int),
		S: new(big.Int).SetBytes(hexutils.HexToBytes("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")),
	})))
	require.False(t, IsInternal(types.NewTx(&types.LegacyTx{
		V: big.NewInt(1),
		R: big.NewInt(1),
		S: big.NewInt(1),
	})))
	require.False(t, IsInternal(types.NewTx(&types.DynamicFeeTx{
		V: big.NewInt(1),
		R: big.NewInt(1),
		S: big.NewInt(1),
	})))
	require.False(t, IsInternal(types.NewTx(&types.LegacyTx{
		V: big.NewInt(1),
		R: new(big.Int),
		S: new(big.Int),
	})))
	require.False(t, IsInternal(types.NewTx(&types.LegacyTx{
		V: new(big.Int),
		R: big.NewInt(1),
		S: new(big.Int),
	})))
}

func TestInternalSender(t *testing.T) {
	require.Equal(t, common.Address{}, InternalSender(types.NewTx(&types.LegacyTx{
		V: new(big.Int),
		R: new(big.Int),
		S: new(big.Int),
	})))
	example := common.HexToAddress("0x0000000000000000000000000000000000000001")
	require.Equal(t, example, InternalSender(types.NewTx(&types.LegacyTx{
		V: new(big.Int),
		R: new(big.Int),
		S: new(big.Int).SetBytes(example.Bytes()),
	})))
	example = common.HexToAddress("0x0000000000000000000000000000000000000100")
	require.Equal(t, example, InternalSender(types.NewTx(&types.LegacyTx{
		V: new(big.Int),
		R: new(big.Int),
		S: new(big.Int).SetBytes(example.Bytes()),
	})))
	example = common.HexToAddress("0x1000000000000000000000000000000000000000")
	require.Equal(t, example, InternalSender(types.NewTx(&types.LegacyTx{
		V: new(big.Int),
		R: new(big.Int),
		S: new(big.Int).SetBytes(example.Bytes()),
	})))
	example = common.HexToAddress("0x1000000000000000000000000000000000000001")
	require.Equal(t, example, InternalSender(types.NewTx(&types.LegacyTx{
		V: new(big.Int),
		R: new(big.Int),
		S: new(big.Int).SetBytes(example.Bytes()),
	})))
}
