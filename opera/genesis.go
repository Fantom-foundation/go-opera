package opera

import "github.com/Fantom-foundation/lachesis-base/hash"

type Genesis struct {
	Rules Rules
	State GenesisState
	Hash  func() hash.Hash
}
