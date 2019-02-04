package posposet

type Store struct {
}

func NewInmemStore() *Store {
	return &Store{}
}
