package pos

import (
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func TestNewValidators(t *testing.T) {
	v := NewValidators()

	assert.NotNil(t, v)
	assert.NotNil(t, v.list)
	assert.NotNil(t, v.indexes)

	assert.Equal(t, 0, v.Len())
}

func TestValidators_Set(t *testing.T) {
	v := NewValidators()

	v.Set(1, 1)
	v.Set(2, 2)
	v.Set(3, 3)
	v.Set(4, 4)
	v.Set(5, 5)

	assert.Equal(t, 5, v.Len())
	assert.Equal(t, Stake(15), v.TotalStake())

	v.Set(1, 10)
	v.Set(3, 30)

	assert.Equal(t, 5, v.Len())
	assert.Equal(t, Stake(51), v.TotalStake())

	v.Set(2, 0)
	v.Set(5, 0)

	assert.Equal(t, 3, v.Len())
	assert.Equal(t, Stake(44), v.TotalStake())

	v.Set(4, 0)
	v.Set(3, 0)
	v.Set(1, 0)

	assert.Equal(t, 0, v.Len())
	assert.Equal(t, Stake(0), v.TotalStake())
}

func TestValidators_Get(t *testing.T) {
	v := NewValidators()

	v.Set(0, 1)
	v.Set(2, 2)
	v.Set(3, 3)
	v.Set(4, 4)
	v.Set(7, 5)

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
	v := NewValidators()

	v.Set(1, 1)
	v.Set(2, 2)
	v.Set(3, 3)
	v.Set(4, 4)
	v.Set(5, 5)

	count := 0
	sum := 0
	for addr := range v.Iterate() {
		count++
		sum += int(v.Get(addr))
	}

	assert.Equal(t, 5, count)
	assert.Equal(t, 15, sum)
}

func TestValidators_Copy(t *testing.T) {
	v := NewValidators()

	v.Set(1, 1)
	v.Set(2, 2)
	v.Set(3, 3)
	v.Set(4, 4)
	v.Set(5, 5)

	vv := v.Copy()

	assert.Equal(t, v.list, vv.list)
	assert.Equal(t, v.indexes, vv.indexes)
	assert.Equal(t, v.ids, vv.ids)

	assert.NotEqual(t, unsafe.Pointer(&v.list), unsafe.Pointer(&vv.list))
	assert.NotEqual(t, unsafe.Pointer(&v.indexes), unsafe.Pointer(&vv.indexes))
	assert.NotEqual(t, unsafe.Pointer(&v.ids), unsafe.Pointer(&vv.ids))
}
