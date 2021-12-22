package bitmap

type Set []byte

func New(max int) Set {
	l := max / 8
	if max%8 != 0 {
		l++
	}
	return make(Set, l)
}

func (s Set) Put(i int) {
	yi := i / 8
	bi := i % 8
	s[yi] |= 1 << bi
}

func (s Set) Del(i int) {
	yi := i / 8
	bi := i % 8
	s[yi] &= ^(1 << bi)
}

func (s Set) Has(i int) bool {
	yi := i / 8
	bi := i % 8
	return (s[yi] & (1 << bi)) != 0
}
