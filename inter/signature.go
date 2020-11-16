package inter

const SigSize = 64

// Signature is a secp256k1 in R|S format
type Signature [SigSize]byte

func (s Signature) Bytes() []byte {
	return s[:]
}

func BytesToSignature(b []byte) (sig Signature) {
	if len(b) != SigSize {
		panic("invalid signature length")
	}
	copy(sig[:], b)
	return sig
}
