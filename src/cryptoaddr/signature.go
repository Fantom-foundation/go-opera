package cryptoaddr

import (
	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	"github.com/Fantom-foundation/go-lachesis/src/hash"
)

// Sign calculates an ECDSA signature.
//
// This function is susceptible to chosen plaintext attacks that can leak
// information about the private key that is used for signing. Callers must
// be aware that the given hash cannot be chosen by an adversery. Common
// solution is to hash any input before calculating the signature.
//
// The produced signature is in the [R || S || V] format where V is 0 or 1.
func Sign(hashToSign hash.Hash, prv *crypto.PrivateKey) ([]byte, error) {
	return crypto.Sign(hashToSign.Bytes(), prv)
}

// VerifySignature returns true if signature was created by a user with this addr.
func VerifySignature(address hash.Peer, signedHash hash.Hash, sig []byte) bool {
	actualAddress, err := RecoverAddr(signedHash, sig)
	if err != nil {
		return false
	}

	return actualAddress == address
}

// RecoverAddr returns the hash of a public key that created the given signature.
func RecoverAddr(signedHash hash.Hash, sig []byte) (hash.Peer, error) {
	pk, err := crypto.SigToPub(signedHash.Bytes(), sig)
	if err != nil {
		return hash.Peer{}, err
	}

	return AddressOf(pk), err
}
