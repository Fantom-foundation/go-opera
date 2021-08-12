package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPrettyDuration_String(t *testing.T) {
	require.Equal(t, "0s", PrettyDuration(0).String())
	require.Equal(t, "1ns", PrettyDuration(time.Nanosecond).String())
	require.Equal(t, "1µs", PrettyDuration(time.Microsecond).String())
	require.Equal(t, "1ms", PrettyDuration(time.Millisecond).String())
	require.Equal(t, "1s", PrettyDuration(time.Second).String())
	require.Equal(t, "1.000s", PrettyDuration(time.Second + time.Microsecond + time.Nanosecond).String())
	require.Equal(t, "1.001s", PrettyDuration(time.Second + time.Millisecond + time.Microsecond + time.Nanosecond).String())
	require.Equal(t, "1m1.001s", PrettyDuration(time.Minute + time.Second + time.Millisecond + time.Microsecond + time.Nanosecond).String())
	require.Equal(t, "1h1m1.001s", PrettyDuration(time.Hour + time.Minute + time.Second + time.Millisecond + time.Microsecond + time.Nanosecond).String())
	require.Equal(t, "1d1h1m", PrettyDuration(24 * time.Hour + time.Hour + time.Minute + time.Second + time.Millisecond + time.Microsecond + time.Nanosecond).String())
	require.Equal(t, "1mo1d1h", PrettyDuration(30 * 24 * time.Hour + 24 * time.Hour + time.Hour + time.Minute + time.Second + time.Millisecond + time.Microsecond + time.Nanosecond).String())
	require.Equal(t, "1y4mo3w", PrettyDuration(503.123456789 *  24 * float64(time.Hour)).String())
}
