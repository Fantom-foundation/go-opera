package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRawSignatureEncoding(t *testing.T) {
	assertar := assert.New(t)
	key, err := GenerateKey()
	if !assertar.NoError(err) {
		return
	}

	msg := "J'aime mieux forger mon ame que la meubler"
	msgBytes := []byte(msg)
	msgHashBytes := Keccak256(msgBytes)

	r, s, err := key.SignRaw(msgHashBytes)
	if !assertar.NoError(err) {
		return
	}

	encodedSig := RawEncodeSignature(r, s)
	dR, dS, err := RawDecodeSignature(encodedSig)
	if !assertar.NoError(err) {
		t.Logf("r: %#v", r)
		t.Logf("s: %#v", s)
		t.Logf("error decoding %v", encodedSig)
		return
	}

	assertar.Equal(r, dR)
	assertar.Equal(s, dS)
}
