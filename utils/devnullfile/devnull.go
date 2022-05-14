package devnullfile

type DevNull struct{}

func (d DevNull) Read(pp []byte) (n int, err error) {
	for i := range pp {
		pp[i] = 0
	}
	return len(pp), nil
}

func (d DevNull) Write(pp []byte) (n int, err error) {
	return len(pp), nil
}

func (d DevNull) Close() error {
	return nil
}

func (d DevNull) Seek(offset int64, whence int) (int64, error) {
	return 0, nil
}

func (d DevNull) Drop() error {
	return nil
}
