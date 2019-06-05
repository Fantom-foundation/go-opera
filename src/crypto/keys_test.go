package crypto

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/src/common"
)

func TestFakeKeyGeneration(t *testing.T) {
	assertar := assert.New(t)

	prev := make([]*common.PrivateKey, 10)
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

func TestKey(t *testing.T) {
	assertar := assert.New(t)
	key := GenerateKey()

	var buf bytes.Buffer
	err := WriteKeyTo(&buf, key)
	if !assertar.NoError(err) {
		return
	}

	nKey, err := ReadPemToKey(&buf)
	if !assertar.NoError(err) {
		return
	}

	assertar.Equal(key, nKey)
}
