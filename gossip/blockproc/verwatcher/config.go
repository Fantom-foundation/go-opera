package verwatcher

import "time"

type Config struct {
	ShutDownIfNotUpgraded     bool
	WarningIfNotUpgradedEvery time.Duration
}
