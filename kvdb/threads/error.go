package threads

import (
	"errors"
)

var (
	ErrExpiredIterator = errors.New("can't open new iterator")
)

// expiredIterator implements empty and failed ethdb.Iterator.
type expiredIterator struct {
}

// Next moves the iterator to the next key/value pair. It returns whether the
// iterator is exhausted.
func (*expiredIterator) Next() bool {
	return false
}

// Error returns any accumulated error. Exhausting all the key/value pairs
// is not considered to be an error.
func (*expiredIterator) Error() error {
	return ErrExpiredIterator
}

// Key returns the key of the current key/value pair, or nil if done. The caller
// should not modify the contents of the returned slice, and its contents may
// change on the next call to Next.
func (*expiredIterator) Key() []byte {
	return nil
}

// Value returns the value of the current key/value pair, or nil if done. The
// caller should not modify the contents of the returned slice, and its contents
// may change on the next call to Next.
func (*expiredIterator) Value() []byte {
	return nil
}

// Release releases associated resources. Release should always succeed and can
// be called multiple times without causing error.
func (*expiredIterator) Release() {
	return
}
