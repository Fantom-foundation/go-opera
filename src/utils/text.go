package utils

import "github.com/satori/go.uuid"

// NewUUID returns UUID as string
func NewUUID() string {
	v := uuid.NewV4()
	return v.String()
}
