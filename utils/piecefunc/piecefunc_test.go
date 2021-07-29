package piecefunc

import (
	"math"
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
		{
			X: math.MaxUint32 * DecimalUnit,
			Y: 1.0 * DecimalUnit,
		},
	})

	for i, x := range []uint64{0, 1, 5, 10, 20, 0xFFFF, math.MaxUint32} {
		X := x * DecimalUnit
		Y := constF(X)
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
		{
			X: math.MaxUint32 * DecimalUnit,
			Y: 21.0 * DecimalUnit,
		},
	})

	for i, x := range []uint64{0, 1, 3, 5, 10, 20} {
		X := x * DecimalUnit
		Y := up45F(X)
		require.Equal(X, Y, i)
	}

	for i, x := range []uint64{21, 0xFFFF, math.MaxUint32} {
		X := x * DecimalUnit
		Y := up45F(X)
		require.True(20.0*DecimalUnit <= Y && Y <= 21.0*DecimalUnit, i)
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
		Y := down45F(X)
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
