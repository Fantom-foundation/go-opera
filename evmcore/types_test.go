package evmcore

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/stretchr/testify/require"
)

func TestStateWrapper(t *testing.T) {
	require := require.New(t)

	mpt, err := state.New(common.Hash{}, state.NewDatabase(rawdb.NewMemoryDatabase()), nil)
	require.NoError(err)
	require.NotNil(mpt)

	wrapped := ToStateDB(mpt)
	require.NotNil(wrapped)

	unwrapped, ok := IsMptStateDB(wrapped)
	require.True(ok)
	require.NotNil(unwrapped)
}
