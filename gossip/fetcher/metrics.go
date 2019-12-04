package fetcher

import (
	"github.com/ethereum/go-ethereum/metrics"
)

var (
	propAnnounceInMeter  = metrics.NewRegisteredGauge("fetcher/prop/announces/in", nil)
	propAnnounceDOSMeter = metrics.NewRegisteredGauge("fetcher/prop/announces/dos", nil)

	propBroadcastInMeter = metrics.NewRegisteredGauge("fetcher/prop/broadcasts/in", nil)

	eventFetchMeter = metrics.NewRegisteredGauge("fetcher/fetch/headers", nil)
)
