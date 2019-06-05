package crypto

import (
	"testing"
)

func TestSignatureEncoding(t *testing.T) {
	privKey, _ := GenerateKey()

	msg := "J'aime mieux forger mon ame que la meubler"
	msgBytes := []byte(msg)
	msgHashBytes := Keccak256(msgBytes)

	r, s, _ := privKey.Sign(msgHashBytes)

	encodedSig := EncodeSignature(r, s)

	dr, ds, err := DecodeSignature(encodedSig)
	if err != nil {
		t.Logf("r: %#v", r)
		t.Logf("s: %#v", s)
		t.Logf("error decoding %v", encodedSig)
		t.Fatal(err)
	}

	if r.Cmp(dr) != 0 {
		t.Fatalf("Signature Rs defer")
	}

	if s.Cmp(ds) != 0 {
		t.Fatalf("Signature Ss defer")
	}

}
