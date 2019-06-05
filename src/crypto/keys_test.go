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

	nKey, err := ReadPemToKey(&buf)
	if !assertar.NoError(err) {
		return
	}

	assertar.Equal(key, nKey)
}
