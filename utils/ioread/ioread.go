package ioread

import "io"

func ReadAll(reader io.Reader, buf []byte) error {
	consumed := 0
	for {
		n, err := reader.Read(buf[consumed:])
		consumed += n
		if consumed == len(buf) {
			return nil
		}
		if err != nil {
			return err
		}
	}
}
