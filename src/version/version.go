package version

import "strings"

// Maj major semver version
const Maj = "0"

// Min minor semver version
const Min = "4"

// Fix fix semver version
const Fix = "5-rc1"

func dashPrependAndSliceOn(condition bool, s string) string {
	if !condition {
		return s
	}
	return "-" + s[:8]
}

var (
	// GitCommit is set with: -ldflags "-X main.gitCommit=$(git rev-parse HEAD)"
	GitCommit string

	// Version the full version string
	Version = strings.Join([]string{Maj, Min, Fix}, ".") + dashPrependAndSliceOn(GitCommit != "", GitCommit)
)
