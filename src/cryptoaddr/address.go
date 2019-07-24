package cryptoaddr

import (
	"github.com/btcsuite/btcd/btcec"

	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	"github.com/Fantom-foundation/go-lachesis/src/hash"
)

// AddressOf calculates hash of the PublicKey.
func AddressOf(pk *crypto.PublicKey) hash.Peer {
	bytes := (*btcec.PublicKey)(pk).SerializeUncompressed()
	return hash.Peer(crypto.Keccak256Hash(bytes))
}
