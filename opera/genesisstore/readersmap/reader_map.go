package readersmap

import (
	"errors"
	"io"
)

type Map map[string]io.Reader

type Unit struct {
	Name string
	io.Reader
}

var (
	ErrNotFound = errors.New("not found")
	ErrDupFile  = errors.New("unit name is duplicated")
)

func Wrap(rr []Unit) (Map, error) {
	units := make(Map)
	for _, r := range rr {
		if units[r.Name] != nil {
			return nil, ErrDupFile
		}
		units[r.Name] = r.Reader
	}
	return units, nil
}

func (mm Map) Open(name string) (io.Reader, error) {
	f := mm[name]
	if f == nil {
		return nil, ErrNotFound
	}
	return f, nil
}
