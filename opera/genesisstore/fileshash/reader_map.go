package fileshash

import (
	"io"

	"github.com/Fantom-foundation/lachesis-base/hash"
)

type Map struct {
	backend func(string) (io.ReadCloser, error)
}

func Wrap(backend func(string) (io.ReadCloser, error), maxMemoryUsage uint64, roots map[string]hash.Hash) func(string) (io.ReadCloser, error) {
	return func(name string) (io.ReadCloser, error) {
		root, ok := roots[name]
		if !ok {
			return nil, ErrRootNotFound
		}
		f, err := backend(name)
		if err != nil {
			return nil, err
		}
		return WrapReader(f, maxMemoryUsage, root), nil
	}
}
