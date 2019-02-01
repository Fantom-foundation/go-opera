package node

import (
	"sync/atomic"
)

type count64 int64

func (c *count64) increment() int64 {
	return atomic.AddInt64((*int64)(c), 1)
}

func (c *count64) decrement() int64 {
	return atomic.AddInt64((*int64)(c), -1)
}

func (c *count64) get() int64 {
	return atomic.LoadInt64((*int64)(c))
}
