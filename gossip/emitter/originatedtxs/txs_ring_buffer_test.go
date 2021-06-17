package originatedtxs

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/common"
)

func TestBuffer(t *testing.T) {
	require := require.New(t)

	addr := make([]common.Address, 50)
	for i := range addr {
		addr[i] = fakeAddr()
	}

	buf := New(len(addr) / 2)

	for _, a := range addr {
		buf.Dec(a)
	}
	for _, a := range addr {
		require.Zero(
			buf.TotalOf(a))
	}
	require.True(
		buf.Empty())

	buf.Clear()
	require.True(
		buf.Empty())

	for i := range addr {
		j := (i + 1) % len(addr)
		buf.Inc(addr[i])
		buf.Inc(addr[j])
	}
	for i, a := range addr {
		total := buf.TotalOf(a)
		switch {
		case i == 0:
			require.Equal(1, total, i)
		case i > len(addr)/2:
			require.Equal(2, total, i)
		default:
			require.Equal(0, total, i)
		}
	}

	for _, a := range addr {
		buf.Dec(a)
		buf.Dec(a)
	}
	require.True(
		buf.Empty())
}

func fakeAddr() (a common.Address) {
	_, err := rand.Read(a[:])
	if err != nil {
		panic(err)
	}
	return
}
