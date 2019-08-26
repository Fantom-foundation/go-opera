package cryptoaddr

import (
	"github.com/btcsuite/btcd/btcec"
	"github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/go-lachesis/src/crypto"
)

// AddressOf calculates hash of the PublicKey.
func AddressOf(pk *crypto.PublicKey) common.Address {
	bytes := (*btcec.PublicKey)(pk).SerializeUncompressed()
	return common.BytesToAddress(crypto.Keccak256(bytes))
}
