package metrics

import "github.com/prometheus/client_golang/prometheus"

type SummaryOpts prometheus.SummaryOpts
type CounterOpts prometheus.CounterOpts
type GaugeOpts prometheus.GaugeOpts
type HistogramOpts prometheus.HistogramOpts

type Vec[T any] interface {
	WithLabelValues(lvs ...string) T
}

type Counter interface {
	Inc()
	Add(float64)
}

type noopCounter struct{}

func (counter noopCounter) Inc()        {}
func (counter noopCounter) Add(float64) {}

type counterVec struct {
	v *prometheus.CounterVec
}

func (c counterVec) WithLabelValues(lvls ...string) Counter {
	return c.v.WithLabelValues(lvls...)
}

func (counter noopCounter) WithLabelValues(lvls ...string) Counter { return noopCounter{} }

type Gauge interface {
	Set(float64)
	Inc()
	Dec()
	Add(float64)
	Sub(float64)
}

type noopGauge struct{}

func (counter noopGauge) Set(float64)                          {}
func (counter noopGauge) Inc()                                 {}
func (counter noopGauge) Dec()                                 {}
func (counter noopGauge) Add(float64)                          {}
func (counter noopGauge) Sub(float64)                          {}
func (counter noopGauge) WithLabelValues(lvls ...string) Gauge { return noopGauge{} }

type gaugeVec struct {
	v *prometheus.GaugeVec
}

func (c gaugeVec) WithLabelValues(lvls ...string) Gauge {
	return c.v.WithLabelValues(lvls...)
}

type Histogram interface {
	Observe(float64)
}

type noopHistogram struct{}

func (counter noopHistogram) Observe(float64)                          {}
func (counter noopHistogram) WithLabelValues(lvls ...string) Histogram { return noopHistogram{} }

type histogramVec struct {
	v *prometheus.HistogramVec
}

func (c histogramVec) WithLabelValues(lvls ...string) Histogram {
	return c.v.WithLabelValues(lvls...)
}

type Summary interface {
	Observe(float64)
}

type noopSummary struct{}

func (counter noopSummary) Observe(float64)                        {}
func (counter noopSummary) WithLabelValues(lvls ...string) Summary { return noopHistogram{} }

type summaryVec struct {
	v *prometheus.SummaryVec
}

func (c summaryVec) WithLabelValues(lvls ...string) Summary {
	return c.v.WithLabelValues(lvls...)
}
