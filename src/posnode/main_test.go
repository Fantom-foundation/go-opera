package posnode

import (
	"io/ioutil"
	"os"
	"testing"
)

const (
	// disables logs for tests
	noLogOutput = true
)

// TestMain is for test mode settings
func TestMain(m *testing.M) {
	if noLogOutput {
		log.Out = ioutil.Discard
	}

	os.Exit(m.Run())
}
