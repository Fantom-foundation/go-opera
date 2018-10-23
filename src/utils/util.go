package utils

import "github.com/satori/go.uuid"

func UUID() string {
	u, err := uuid.NewV4()
	if err != nil {
		return ""
	}
	return u.String()
}
