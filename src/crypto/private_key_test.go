package crypto

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrivateKey(t *testing.T) {
	assertar := assert.New(t)
	key, err := GeneratePrivateKey()
	if !assertar.NoError(err) {
		return
	}

	var buf bytes.Buffer
	err = key.WriteTo(&buf)
	if !assertar.NoError(err) {
		return
	}

	nKey, err := ReadPrivateKey(&buf)
	if !assertar.NoError(err) {
		return
	}

	assertar.Equal(key, nKey)
}
