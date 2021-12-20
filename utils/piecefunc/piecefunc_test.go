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
			X: 0xFFFFFF * DecimalUnit,
			Y: 1.0 * DecimalUnit,
		},
	})

	for i, x := range []uint64{0, 1, 5, 10, 20, 0xFFFF, 0xFFFFFF} {
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
			X: 0xFFFFFF * DecimalUnit,
			Y: 21.0 * DecimalUnit,
		},
	})

	for i, x := range []uint64{0, 1, 3, 5, 10, 20} {
		X := x * DecimalUnit
		Y := up45F(X)
		require.Equal(X, Y, i)
	}

	for i, x := range []uint64{21, 0xFFFF, 0xFFFFFF} {
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
	require.Panics(func() {
		// too large val
		NewFunc([]Dot{
			{
				X: 0,
				Y: 0,
			},
			{
				X: maxVal + 1,
				Y: 1,
			},
		})
	})
	require.Panics(func() {
		// too large val
		NewFunc([]Dot{
			{
				X: 0,
				Y: 0,
			},
			{
				X: 1,
				Y: maxVal + 1,
			},
		})
	})
	require.Panics(func() {
		// too large val
		NewFunc([]Dot{
			{
				X: maxVal + 1,
				Y: maxVal + 1,
			},
			{
				X: maxVal + 2,
				Y: maxVal + 2,
			},
		})
	})
}

func TestCustomFunc(t *testing.T) {
	require := require.New(t)

	var f = NewFunc([]Dot{
		{
			X: 1,
			Y: 0.0 * DecimalUnit,
		},
		{
			X: 10.0 * DecimalUnit,
			Y: 10.0 * DecimalUnit,
		},
		{
			X: 20.0 * DecimalUnit,
			Y: 40.0 * DecimalUnit,
		},
		{
			X: 25.0 * DecimalUnit,
			Y: 41.0 * DecimalUnit,
		},
		{
			X: maxVal,
			Y: maxVal,
		},
	})

	require.Equal(uint64(0), f(0))
	require.Equal(uint64(0), f(1))
	require.Equal(uint64(0), f(10))
	require.Equal(uint64(10), f(11))
	require.Equal(uint64(4.99999*DecimalUnit), f(5.0*DecimalUnit))
	require.Equal(uint64(10.0*DecimalUnit), f(10.0*DecimalUnit))
	require.Equal(uint64(25.0*DecimalUnit), f(15.0*DecimalUnit))
	require.Equal(uint64(34.0*DecimalUnit), f(18.0*DecimalUnit))
	require.Equal(uint64(37.0*DecimalUnit), f(19.0*DecimalUnit))
	require.Equal(uint64(37.3*DecimalUnit), f(19.1*DecimalUnit))
	require.Equal(uint64(40.0*DecimalUnit), f(20.0*DecimalUnit))
	require.Equal(uint64(40.4*DecimalUnit), f(22.0*DecimalUnit))
	require.Equal(uint64(41*DecimalUnit), f(25.0*DecimalUnit))
	require.Equal(uint64(250012.273351*DecimalUnit), f(250000.0*DecimalUnit))
	require.Equal(uint64(2500011.987361*DecimalUnit), f(2500000.0*DecimalUnit))
	require.Equal(uint64(maxVal-18446704-18446703), f(maxVal-18446704-16))
	require.Equal(uint64(maxVal-18446704), f(maxVal-18446704-15))
	require.Equal(uint64(maxVal-18446704), f(maxVal-1))
	require.Equal(uint64(maxVal), f(maxVal))
	require.Equal(uint64(maxVal), f(math.MaxUint64))
}
