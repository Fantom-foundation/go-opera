package migration

import (
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/log"
	"golang.org/x/crypto/sha3"
)

// Migration is a migration step.
type Migration struct {
	name string
	exec func() error
	prev *Migration
}

// Begin with empty unique migration step.
func Begin(appName string) *Migration {
	return &Migration{
		name: appName,
	}
}

// Next creates next migration.
func (m *Migration) Next(name string, exec func() error) *Migration {
	if name == "" {
		panic("empty name")
	}

	if exec == nil {
		panic("empty exec")
	}

	return &Migration{
		name: name,
		exec: exec,
		prev: m,
	}
}

// ID is an uniq migration's id.
func (m *Migration) ID() string {
	digest := sha3.NewLegacyKeccak256()
	digest.Write([]byte(m.name))

	bytes := digest.Sum(nil)
	return fmt.Sprintf("%x", bytes)
}

// Exec method run migrations chain in order
func (m *Migration) Exec(curr IDStore) error {
	currID := curr.GetID()
	myID := m.ID()

	if m.veryFirst() {
		if currID != "" {
			return errors.New("unknown version: " + currID)
		}
		return nil
	}

	if currID == myID {
		return nil
	}

	err := m.prev.Exec(curr)
	if err != nil {
		return err
	}

	err = m.exec()
	if err != nil {
		log.Error("'"+m.name+"' migration failed", "err", err)
		return err
	}
	log.Warn("'" + m.name + "' migration has been applied")

	curr.SetID(myID)
	return nil
}

func (m *Migration) veryFirst() bool {
	return m.exec == nil
}

// IDs return list of migrations ids in chain
func (m *Migration) IDs() []string {
	if m.prev == nil {
		return []string{m.ID()}
	}

	return append(m.prev.IDs(), m.ID())
}
