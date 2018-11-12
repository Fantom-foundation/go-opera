package difftool

// ByValue implements sort.Interface for []string
type ByValue []string

func (a ByValue) Len() int { return len(a) }

func (a ByValue) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

func (a ByValue) Less(i, j int) bool {
	s0 := a[i]
	s1 := a[j]

	if len(s0) < len(s0) {
		return true
	}
	if len(s0) > len(s1) {
		return false
	}

	for n := 0; n < len(s0); n++ {
		if s0[n] < s1[n] {
			return true
		}
		if s0[n] > s1[n] {
			return false
		}
	}

	return false
}
