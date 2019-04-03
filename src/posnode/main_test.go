package posnode

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// Disable logs in test mode
	log.Out = ioutil.Discard
	os.Exit(m.Run())
}
