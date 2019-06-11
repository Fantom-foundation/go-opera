package metrics

import (
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Fantom-foundation/go-lachesis/src/logger"
)

var (
	// Enabled flag for turn on/off metrics
	Enabled = false

	log = logger.Get().WithField("module", "metrics")
)

const (
	envEnabled = "METRICS_ENABLED"
)

func init() {
	for _, arg := range os.Args {
		flag := strings.TrimLeft(arg, "-")

		if flag == "metrics" {
			Enabled = true
			log.Debug("metrics is enabled by flags")
			return
		}
	}

	isEnabled, ok := os.LookupEnv(envEnabled)
	if ok {
		switch strings.ToLower(isEnabled) {
		case "1", "true", "on":
			Enabled = true
			log.Debug("metrics is enabled by environment")
		case "0", "false", "off":
			Enabled = false
			log.Debug("metrics is disable by environment")
		default:
			log.Errorf("incorrect value in '%s'", envEnabled)
		}
	}
}

// Metric property measure with time attributes
type Metric interface {
	// CreationTime return creation time of object implemented Metric
	CreationTime() time.Time

	// LastModification return modification time (update after changing metric property)
	LastModification() time.Time

	// updateModification change modification time
	updateModification()

	// cope make snapshot
	copy() Metric
}

func newStandardMetric(loc *time.Location) Metric {
	if !Enabled {
		return &nilMetric{}
	}

	if loc == nil {
		loc = time.UTC
	}

	currentTime := time.Now().In(loc)
	return &standardMetric{
		loc:              loc,
		creationTime:     currentTime,
		lastModification: currentTime,
	}
}

type standardMetric struct {
	mu sync.RWMutex

	loc              *time.Location
	creationTime     time.Time
	lastModification time.Time
}

func (m *standardMetric) CreationTime() time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.creationTime
}

func (m *standardMetric) LastModification() time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.lastModification
}

func (m *standardMetric) updateModification() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.lastModification = time.Now().In(m.loc)
}

func (m *standardMetric) copy() Metric {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return newStandardMetric(m.loc)
}

type nilMetric struct{}

func (*nilMetric) CreationTime() time.Time {
	return time.Unix(0, 0).UTC()
}

func (*nilMetric) LastModification() time.Time {
	return time.Unix(0, 0).UTC()
}

func (*nilMetric) updateModification() {}

func (*nilMetric) copy() Metric {
	return &nilMetric{}
}
