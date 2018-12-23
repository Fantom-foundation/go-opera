package common

import "fmt"

// StoreErrType store error type
type StoreErrType uint32

const (
	// KeyNotFound PEM key not found
	KeyNotFound StoreErrType = iota
	// TooLate TODO
	TooLate
	// PassedIndex passed the index lookup
	PassedIndex
	// SkippedIndex skipped the index lookup
	SkippedIndex
	// NoRoot no root to load
	NoRoot
	// UnknownParticipant signed participant is not known
	UnknownParticipant
	// Empty TODO
	Empty
	// KeyAlreadyExists key already exists in the store
	KeyAlreadyExists
)

// StoreErr storage error
type StoreErr struct {
	dataType string
	errType  StoreErrType
	key      string
}

// NewStoreErr constructor
func NewStoreErr(dataType string, errType StoreErrType, key string) StoreErr {
	return StoreErr{
		dataType: dataType,
		errType:  errType,
		key:      key,
	}
}

func (e StoreErr) Error() string {
	m := ""
	switch e.errType {
	case KeyNotFound:
		m = "Not Found"
	case TooLate:
		m = "Too Late"
	case PassedIndex:
		m = "Passed Index"
	case SkippedIndex:
		m = "Skipped Index"
	case NoRoot:
		m = "No Root"
	case UnknownParticipant:
		m = "Unknown Participant"
	case Empty:
		m = "Empty"
	}

	return fmt.Sprintf("%s, %s, %s", e.dataType, e.key, m)
}

// Is checks if store error type
func Is(err error, t StoreErrType) bool {
	storeErr, ok := err.(StoreErr)
	return ok && storeErr.errType == t
}
