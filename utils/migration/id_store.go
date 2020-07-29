package migration

// IDStore interface for stores id in migrations
type IDStore interface {
	GetID() string
	SetID(string)
}

type inmemIDStore struct {
	lastID string
}

func (p *inmemIDStore) GetID() string {
	return string(p.lastID)
}

func (p *inmemIDStore) SetID(id string) {
	p.lastID = id
}
