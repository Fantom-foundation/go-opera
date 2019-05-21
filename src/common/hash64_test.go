package common
import (
	"testing"
)
func TestHash64(t *testing.T) {
	data := []byte{'1', '2', 3, 4, 5}
	h := Hash64(data)
	if h != 5005865916200746024 {
		t.Errorf("Hash64 of test data expected to be 5005865916200746024 but found %v!", h)
	}
}
