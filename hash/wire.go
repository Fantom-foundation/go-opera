package hash

// ToWire converts to simple slice.
func (hh Events) ToWire() [][]byte {
	res := make([][]byte, len(hh))
	for i, h := range hh {
		res[i] = h.Bytes()
	}

	return res
}

// WireToEvents converts from simple slice.
func WireToEvents(buf [][]byte) Events {
	if buf == nil {
		return nil
	}

	hh := make(Events, len(buf))
	for i, b := range buf {
		hh[i] = BytesToEvent(b)
	}

	return hh
}
