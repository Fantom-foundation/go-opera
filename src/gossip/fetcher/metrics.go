package fetcher

import "github.com/Fantom-foundation/go-lachesis/src/metrics"

var (
	propAnnounceInMeter  = metrics.RegisterGauge("fantom/fetcher/prop/announces/in", nil)
	propAnnounceDOSMeter = metrics.RegisterGauge("fantom/fetcher/prop/announces/dos", nil)

	propBroadcastInMeter = metrics.RegisterGauge("fantom/fetcher/prop/broadcasts/in", nil)

	eventFetchMeter = metrics.RegisterGauge("fantom/fetcher/fetch/headers", nil)
)
