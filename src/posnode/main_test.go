package posnode

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/Fantom-foundation/go-lachesis/src/logger"
)

const (
	// disables logs for tests
	noLogOutput = true
)

// TestMain is for test mode settings
func TestMain(m *testing.M) {
	if noLogOutput {
		logger.Get().Out = ioutil.Discard
	}

	os.Exit(m.Run())
}
