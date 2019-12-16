package pos

import (
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func TestNewValidators(t *testing.T) {
	b := NewBuilder()

	assert.NotNil(t, b)
	assert.NotNil(t, b.Build())

	assert.Equal(t, 0, b.Build().Len())
}

func TestValidators_Set(t *testing.T) {
	b := NewBuilder()

	b.Set(1, 1)
	b.Set(2, 2)
	b.Set(3, 3)
	b.Set(4, 4)
	b.Set(5, 5)

	v := b.Build()

	assert.Equal(t, 5, v.Len())
	assert.Equal(t, Stake(15), v.TotalStake())

	b.Set(1, 10)
	b.Set(3, 30)

	v = b.Build()

	assert.Equal(t, 5, v.Len())
	assert.Equal(t, Stake(51), v.TotalStake())

	b.Set(2, 0)
	b.Set(5, 0)

	v = b.Build()

	assert.Equal(t, 3, v.Len())
	assert.Equal(t, Stake(44), v.TotalStake())

	b.Set(4, 0)
	b.Set(3, 0)
	b.Set(1, 0)

	v = b.Build()

	assert.Equal(t, 0, v.Len())
	assert.Equal(t, Stake(0), v.TotalStake())
}

func TestValidators_Get(t *testing.T) {
	b := NewBuilder()

	b.Set(0, 1)
	b.Set(2, 2)
	b.Set(3, 3)
	b.Set(4, 4)
	b.Set(7, 5)

	v := b.Build()

	assert.Equal(t, Stake(1), v.Get(0))
	assert.Equal(t, Stake(0), v.Get(1))
	assert.Equal(t, Stake(2), v.Get(2))
	assert.Equal(t, Stake(3), v.Get(3))
	assert.Equal(t, Stake(4), v.Get(4))
	assert.Equal(t, Stake(0), v.Get(5))
	assert.Equal(t, Stake(0), v.Get(6))
	assert.Equal(t, Stake(5), v.Get(7))
}

func TestValidators_Iterate(t *testing.T) {
	b := NewBuilder()

	b.Set(1, 1)
	b.Set(2, 2)
	b.Set(3, 3)
	b.Set(4, 4)
	b.Set(5, 5)

	v := b.Build()

	count := 0
	sum := 0

	for _, id := range v.IDs() {
		count++
		sum += int(v.Get(id))
	}

	assert.Equal(t, 5, count)
	assert.Equal(t, 15, sum)
}

func TestValidators_Copy(t *testing.T) {
	b := NewBuilder()

	b.Set(1, 1)
	b.Set(2, 2)
	b.Set(3, 3)
	b.Set(4, 4)
	b.Set(5, 5)

	v := b.Build()
	vv := v.Copy()

	assert.Equal(t, v.values, vv.values)

	assert.NotEqual(t, unsafe.Pointer(&v.values), unsafe.Pointer(&vv.values))
	assert.NotEqual(t, unsafe.Pointer(&v.cache.indexes), unsafe.Pointer(&vv.cache.indexes))
	assert.NotEqual(t, unsafe.Pointer(&v.cache.ids), unsafe.Pointer(&vv.cache.ids))
	assert.NotEqual(t, unsafe.Pointer(&v.cache.stakes), unsafe.Pointer(&vv.cache.stakes))
}
