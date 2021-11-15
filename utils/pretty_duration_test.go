package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPrettyDuration_String(t *testing.T) {
	for _, testcase := range []struct {
		str string
		val time.Duration
	}{
		{"0s", 0},
		{"1ns", time.Nanosecond},
		{"1µs", time.Microsecond},
		{"1ms", time.Millisecond},
		{"1s", time.Second},
		{"1.000s", time.Second + time.Microsecond + time.Nanosecond},
		{"1.001s", time.Second + time.Millisecond + time.Microsecond + time.Nanosecond},
		{"1m1.001s", time.Minute + time.Second + time.Millisecond + time.Microsecond + time.Nanosecond},
		{"1h1m1.001s", time.Hour + time.Minute + time.Second + time.Millisecond + time.Microsecond + time.Nanosecond},
		{"1d1h1m", 24*time.Hour + time.Hour + time.Minute + time.Second + time.Millisecond + time.Microsecond + time.Nanosecond},
		{"1mo1d1h", 30*24*time.Hour + 24*time.Hour + time.Hour + time.Minute + time.Second + time.Millisecond + time.Microsecond + time.Nanosecond},
		{"26y4mo23d", time.Duration(9503.123456789 * 24 * float64(time.Hour))},
	} {
		require.Equal(t, testcase.str, PrettyDuration(testcase.val).String())
		if testcase.val > 0 {
			require.Equal(t, "-"+testcase.str, PrettyDuration(-testcase.val).String())
		}
	}
}
