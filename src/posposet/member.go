package posposet

import "github.com/Fantom-foundation/go-lachesis/src/crypto"

// Member is a team member.
type Member struct {
	number    int
	publicKey string // string in base64
	stake     uint64 // In units
}

// NewMember creates a new member.
func NewMember(publicKey string, stake uint64) *Member {
	return &Member{
		publicKey: publicKey,
		stake:     stake,
	}
}

// SetPublicKey sets a public key.
func (m *Member) SetPublicKey(key crypto.PublicKey) {
	m.publicKey = key.Base64()
}

// SetAmount sets a stake.
func (m *Member) SetStake(stake uint64) {
	m.stake = stake
}

// Amount gets a member stake.
func (m *Member) Stake() uint64 {
	return m.stake
}

// PublicKey gets a member public key.
func (m *Member) PublicKey() string {
	return m.publicKey
}

// Member gets a member number.
func (m *Member) Number() int {
	return m.number
}
