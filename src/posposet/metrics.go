package posposet

import (
	"github.com/Fantom-foundation/go-lachesis/src/metrics"
)

var (
	// frame cache size in bytes.
	frameCacheSize = metrics.RegisterGauge("frame_cache_size", nil)

	// event to frame cache size in bytes.
	event2FrameCacheSize = metrics.RegisterGauge("event_2_frame_cache_size", nil)

	// event to block cache size in bytes.
	event2BlockCacheSize = metrics.RegisterGauge("event_2_block_cache_size", nil)

	// frame cache capacity.
	frameCacheCap = metrics.RegisterGauge("frame_cache_cap", nil)

	// event to frame cache capacity.
	event2FrameCacheCap = metrics.RegisterGauge("event_2_frame_cache_cap", nil)

	// event to block cache capacity.
	event2BlockCacheCap = metrics.RegisterGauge("event_2_block_cache_cap", nil)
)
