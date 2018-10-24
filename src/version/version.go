package version

import "strings"

const Maj = "0"
const Min = "3"
const Fix = "3"

func dashPrependAndSliceOn(condition bool, s string) string {
	if !condition {
		return s
	}
	return "-" + s[:8]
}

var (
	// GitCommit is set with: -ldflags "-X main.gitCommit=$(git rev-parse HEAD)"
	GitCommit string

	// The full version string
	Version = strings.Join([]string{Maj, Min, Fix}, ".") + dashPrependAndSliceOn(GitCommit != "", GitCommit)
)
