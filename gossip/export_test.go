package gossip

import "time"

func setProgressThreshold(threshold time.Duration) {
	noProgressTime = threshold
}

func setApplicationThreshold(threshold time.Duration) {
	noAppMessageTime = threshold
}
