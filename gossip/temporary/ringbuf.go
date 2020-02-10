package temporary

import (
	"fmt"
)

type ringbuf struct {
	Min    uint64
	offset int
	count  int
	seq    [5]*pair // len is a max count
}

func (r *ringbuf) Get(num uint64) *pair {
	i := int(num - r.Min)
	if num < r.Min || i >= r.count {
		return nil
	}

	i = (i + r.offset) % len(r.seq)
	return r.seq[i]
}

func (r *ringbuf) Set(num uint64, val *pair) {
	if r.count >= len(r.seq) {
		panic("no space")
	}

	if r.count == 0 {
		r.Min = num
	}

	i := int(num - r.Min)
	if i != r.count {
		panic(fmt.Sprintf("sequence is broken (set %d to %+v)", num, r))
	}

	i = (i + r.offset) % len(r.seq)
	r.seq[i] = val
	r.count++
}

func (r *ringbuf) Del(num uint64) {
	if num != r.Min {
		panic(fmt.Sprintf("sequence is broken (del %d from %+v)", num, r))
	}

	r.Min++
	r.offset = (r.offset + 1) % len(r.seq)
	r.count--
}
