package poset

import (
	"github.com/ethereum/go-ethereum/metrics"
)

var (
	// frame cache capacity.
	frameCacheCap = metrics.NewRegisteredGauge("poset/frame_cache_cap", nil)

	// event to frame cache capacity.
	event2FrameCacheCap = metrics.NewRegisteredGauge("poset/event_2_frame_cache_cap", nil)

	// event to block cache capacity.
	event2BlockCacheCap = metrics.NewRegisteredGauge("poset/event_2_block_cache_cap", nil)
)
