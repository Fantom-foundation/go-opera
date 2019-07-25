package crypto

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFakeKeyGeneration(t *testing.T) {
	assertar := assert.New(t)

	prev := make([]*PrivateKey, 10)
	for i := 0; i < len(prev); i++ {
		prev[i] = GenerateFakeKey(i)
	}

	for i := 0; i < len(prev); i++ {
		again := GenerateFakeKey(i)
		if !assertar.Equal(prev[i], again) {
			return
		}
	}
}

func TestPrivateKeySignVerify(t *testing.T) {
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

	assertar.True(key.Public().VerifyRaw(msgHashBytes, r, s))
}

func TestPublicBytes(t *testing.T) {
	assertar := assert.New(t)
	key, err := GenerateKey()
	if !assertar.NoError(err) {
		return
	}

	public := key.Public()
	bytes := public.Bytes()
	rKey := BytesToPubKey(bytes)

	assertar.Equal(public, rKey)
}

func TestPublicBase64(t *testing.T) {
	assertar := assert.New(t)
	key, err := GenerateKey()
	if !assertar.NoError(err) {
		return
	}

	public := key.Public()
	base64 := public.Base64()
	rKey, err := Base64ToPubKey(base64)
	if !assertar.NoError(err) {
		return
	}

	assertar.Equal(public, rKey)
}

func TestKeyGenWriteRead(t *testing.T) {
	assertar := assert.New(t)
	key, err := GenerateKey()
	if !assertar.NoError(err) {
		return
	}

	var buf bytes.Buffer
	err = key.WriteTo(&buf)
	if !assertar.NoError(err) {
		return
	}

	nKey, err := ReadKey(&buf)
	if !assertar.NoError(err) {
		return
	}

	assertar.Equal(key, nKey)
}
