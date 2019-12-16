package prometheus

import (
	"strconv"

	"github.com/ethereum/go-ethereum/metrics"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

// NewCollector constructor.
func NewCollector(opts prometheus.Opts, metric interface{}, fields ...string) *Collector {

	return &Collector{
		Metric: &Metric{
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(opts.Namespace, opts.Subsystem, opts.Name),
				"",
				fields,
				nil,
			),
			m: metric,
		},
	}
}

// Collector collects ethereum metrics data.
type Collector struct {
	*Metric
}

// Describe implements prometheus.Collector interface.
func (c *Collector) Describe(out chan<- *prometheus.Desc) {
	out <- c.Metric.Desc()
}

// Collect implements prometheus.Collector interface.
func (c *Collector) Collect(out chan<- prometheus.Metric) {
	out <- c.Metric
}

type Metric struct {
	desc *prometheus.Desc
	m    interface{}
}

func (m *Metric) Desc() *prometheus.Desc {
	return m.desc
}

func (m *Metric) Write(out *dto.Metric) error {
	switch metric := m.m.(type) {

	case metrics.Meter:
		t := metric.Snapshot()

		sum := &dto.Summary{
			SampleCount: new(uint64),
		}
		*sum.SampleCount = uint64(t.Count())

		out.Summary = sum
		out.Label = []*dto.LabelPair{
			pairF("rate1m", t.Rate1()),
			pairF("rate5m", t.Rate5()),
			pairF("rate15m", t.Rate15()),
			pairF("rate", t.RateMean()),
		}

	case metrics.Histogram:
		t := metric.Snapshot()

		sum := &dto.Summary{
			SampleCount: new(uint64),
			SampleSum:   new(float64),
		}
		*sum.SampleCount = uint64(t.Count())
		*sum.SampleSum = float64(t.Sum())

		qq := []float64{0.5, 0.75, 0.95, 0.99, 0.999}
		ps := t.Percentiles(qq)

		for i := range qq {
			sum.Quantile = append(sum.Quantile, quantile(qq[i], ps[i]))
		}

		out.Summary = sum
		out.Label = []*dto.LabelPair{
			pairI("min", t.Min()),
			pairI("max", t.Max()),
			pairF("mean", t.Mean()),
		}

	case metrics.Timer:
		t := metric.Snapshot()

		sum := &dto.Summary{
			SampleCount: new(uint64),
			SampleSum:   new(float64),
		}
		*sum.SampleCount = uint64(t.Count())
		*sum.SampleSum = float64(t.Sum())

		qq := []float64{0.5, 0.75, 0.95, 0.99, 0.999}
		ps := t.Percentiles(qq)

		for i := range qq {
			sum.Quantile = append(sum.Quantile, quantile(qq[i], ps[i]))
		}

		out.Summary = sum
		out.Label = []*dto.LabelPair{
			pairI("min", t.Min()),
			pairI("max", t.Max()),
			pairF("mean", t.Mean()),
			pairF("rate1m", t.Rate1()),
			pairF("rate5m", t.Rate5()),
			pairF("rate15m", t.Rate15()),
			pairF("rate", t.RateMean()),
		}

	case metrics.ResettingTimer:
		t := metric.Snapshot()

		sum := &dto.Summary{}

		qq := []float64{0.5, 0.75, 0.95, 0.99, 0.999}
		ps := t.Percentiles(qq)

		for i := range qq {
			sum.Quantile = append(sum.Quantile, quantile(qq[i], float64(ps[i])))
		}

		out.Summary = sum
	}
	return nil
}

func quantile(q, v float64) *dto.Quantile {
	return &dto.Quantile{
		Quantile: &q,
		Value:    &v,
	}
}

func pairI(name string, val int64) *dto.LabelPair {
	s := strconv.FormatInt(val, 10)
	return &dto.LabelPair{
		Name:  &name,
		Value: &s,
	}
}

func pairF(name string, val float64) *dto.LabelPair {
	s := strconv.FormatFloat(val, 'g', -1, 64)
	return &dto.LabelPair{
		Name:  &name,
		Value: &s,
	}
}
