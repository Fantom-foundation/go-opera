package fetcher

import "github.com/Fantom-foundation/go-lachesis/src/metrics"

var (
	propAnnounceInMeter  = metrics.RegisterGauge("lachesis/fetcher/prop/announces/in", nil)
	propAnnounceDOSMeter = metrics.RegisterGauge("lachesis/fetcher/prop/announces/dos", nil)

	propBroadcastInMeter = metrics.RegisterGauge("lachesis/fetcher/prop/broadcasts/in", nil)

	eventFetchMeter = metrics.RegisterGauge("lachesis/fetcher/fetch/headers", nil)
)
