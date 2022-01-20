package fileszip

import (
	"archive/zip"
	"errors"
	"io"
)

type Map struct {
	files map[string]*zip.File
}

type Reader struct {
	Reader io.ReaderAt
	Size   int64
}

var (
	ErrNotFound = errors.New("not found")
	ErrDupFile  = errors.New("file is duplicated in multiple zip archives")
)

func Open(rr []Reader) (*Map, error) {
	zips := make([]*zip.Reader, 0, len(rr))
	files := make(map[string]*zip.File)
	for _, r := range rr {
		z, err := zip.NewReader(r.Reader, r.Size)
		if err != nil {
			return nil, err
		}
		for _, f := range z.File {
			if files[f.Name] != nil {
				return nil, ErrDupFile
			}
			files[f.Name] = f
		}
		zips = append(zips, z)
	}
	return &Map{
		files: files,
	}, nil
}

func (r *Map) Open(name string) (io.ReadCloser, uint64, error) {
	f := r.files[name]
	if f == nil {
		return nil, 0, ErrNotFound
	}
	stream, err := f.Open()
	if err != nil {
		return nil, 0, err
	}
	return stream, f.UncompressedSize64, nil
}
