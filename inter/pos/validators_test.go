package pos

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"testing"
	"unsafe"
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

	v.Set(common.Address{1}, 1)
	v.Set(common.Address{1, 2}, 2)
	v.Set(common.Address{1, 2, 3}, 3)
	v.Set(common.Address{1, 2, 3, 4}, 4)
	v.Set(common.Address{1, 2, 3, 4, 5}, 5)

	assert.Equal(t, 5, v.Len())
	assert.Equal(t, Stake(15), v.TotalStake())

	v.Set(common.Address{1}, 10)
	v.Set(common.Address{1, 2, 3}, 30)

	assert.Equal(t, 5, v.Len())
	assert.Equal(t, Stake(51), v.TotalStake())

	v.Set(common.Address{1, 2}, 0)
	v.Set(common.Address{1, 2, 3, 4, 5}, 0)

	assert.Equal(t, 3, v.Len())
	assert.Equal(t, Stake(44), v.TotalStake())

	v.Set(common.Address{1, 2, 3, 4}, 0)
	v.Set(common.Address{1, 2, 3}, 0)
	v.Set(common.Address{1}, 0)

	assert.Equal(t, 0, v.Len())
	assert.Equal(t, Stake(0), v.TotalStake())
}

func TestValidators_Get(t *testing.T) {
	v := NewValidators()

	v.Set(common.Address{1}, 1)
	v.Set(common.Address{1, 2}, 2)
	v.Set(common.Address{1, 2, 3}, 3)
	v.Set(common.Address{1, 2, 3, 4}, 4)
	v.Set(common.Address{1, 2, 3, 4, 5}, 5)

	assert.Equal(t, Stake(1), v.Get(common.Address{1}))
	assert.Equal(t, Stake(2), v.Get(common.Address{1, 2}))
	assert.Equal(t, Stake(3), v.Get(common.Address{1, 2, 3}))
	assert.Equal(t, Stake(4), v.Get(common.Address{1, 2, 3, 4}))
	assert.Equal(t, Stake(5), v.Get(common.Address{1, 2, 3, 4, 5}))
}

func TestValidators_Iterate(t *testing.T) {
	v := NewValidators()

	v.Set(common.Address{1}, 1)
	v.Set(common.Address{1, 2}, 2)
	v.Set(common.Address{1, 2, 3}, 3)
	v.Set(common.Address{1, 2, 3, 4}, 4)
	v.Set(common.Address{1, 2, 3, 4, 5}, 5)

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

	v.Set(common.Address{1}, 1)
	v.Set(common.Address{1, 2}, 2)
	v.Set(common.Address{1, 2, 3}, 3)
	v.Set(common.Address{1, 2, 3, 4}, 4)
	v.Set(common.Address{1, 2, 3, 4, 5}, 5)

	vv := v.Copy()

	assert.Equal(t, v.list, vv.list)
	assert.Equal(t, v.indexes, vv.indexes)
	assert.Equal(t, v.addresses, vv.addresses)

	assert.NotEqual(t, unsafe.Pointer(&v.list), unsafe.Pointer(&vv.list))
	assert.NotEqual(t, unsafe.Pointer(&v.indexes), unsafe.Pointer(&vv.indexes))
	assert.NotEqual(t, unsafe.Pointer(&v.addresses), unsafe.Pointer(&vv.addresses))
}
