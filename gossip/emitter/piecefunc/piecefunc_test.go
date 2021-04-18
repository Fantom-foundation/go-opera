package piecefunc

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConstFunc(t *testing.T) {
	require := require.New(t)

	var constF = NewFunc([]Dot{
		{
			X: 0.0 * DecimalUnit,
			Y: 1.0 * DecimalUnit,
		},
		{
			X: 10.0 * DecimalUnit,
			Y: 1.0 * DecimalUnit,
		},
		{
			X: 20.0 * DecimalUnit,
			Y: 1.0 * DecimalUnit,
		},
	})

	for i, x := range []uint64{0, 1, 5, 10, 20} {
		X := x * DecimalUnit
		Y := Get(X, constF)
		require.Equal(uint64(DecimalUnit), Y, i)
	}
}

func TestUp45Func(t *testing.T) {
	require := require.New(t)

	var up45F = NewFunc([]Dot{
		{
			X: 0.0 * DecimalUnit,
			Y: 0.0 * DecimalUnit,
		},
		{
			X: 10.0 * DecimalUnit,
			Y: 10.0 * DecimalUnit,
		},
		{
			X: 20.0 * DecimalUnit,
			Y: 20.0 * DecimalUnit,
		},
	})

	for i, x := range []uint64{0, 1, 3, 5, 10, 20} {
		X := x * DecimalUnit
		Y := Get(X, up45F)
		require.Equal(X, Y, i)
	}
}

func TestDown45Func(t *testing.T) {
	require := require.New(t)

	var down45F = NewFunc([]Dot{
		{
			X: 0.0 * DecimalUnit,
			Y: 20.0 * DecimalUnit,
		},
		{
			X: 10.0 * DecimalUnit,
			Y: 10.0 * DecimalUnit,
		},
		{
			X: 20.0 * DecimalUnit,
			Y: 0.0 * DecimalUnit,
		},
	})

	for i, x := range []uint64{0, 1, 2, 5, 10, 20} {
		X := x * DecimalUnit
		Y := Get(X, down45F)
		require.Equal(20.0*DecimalUnit-X, Y, i)
	}
}

func TestFuncCheck(t *testing.T) {
	require := require.New(t)

	require.Panics(func() {
		// too few dots
		_ = NewFunc([]Dot{
			{
				X: 0.0 * DecimalUnit,
				Y: 20.0 * DecimalUnit,
			},
		})
	})

	require.Panics(func() {
		// non monotonic X
		_ = NewFunc([]Dot{
			{
				X: 0.0 * DecimalUnit,
				Y: 20.0 * DecimalUnit,
			},
			{
				X: 20.0 * DecimalUnit,
				Y: 20.0 * DecimalUnit,
			},
			{
				X: 10.0 * DecimalUnit,
				Y: 20.0 * DecimalUnit,
			},
		})
	})
}
