package common

// Int64Slice allows sorting of 64-bit integer slices, like PendingRoundReceived
type Int64Slice []int64

func (a Int64Slice) Len() int      { return len(a) }
func (a Int64Slice) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a Int64Slice) Less(i, j int) bool {
	return a[i] < a[j]
}
