package posposet

type superFrame struct {
	frames  map[uint64]*Frame
	members members
}

func newSuperFrame() *superFrame {
	return &superFrame{
		frames:  make(map[uint64]*Frame),
		members: members{},
	}
}
