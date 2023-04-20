package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Factory interface {
	NewCounter(opts CounterOpts) Counter
	NewCounterVec(opts CounterOpts, labelNames []string) Vec[Counter]
	NewGauge(opts GaugeOpts) Gauge
	NewGaugeVec(opts GaugeOpts, labelNames []string) Vec[Gauge]
	NewHistogram(opts HistogramOpts) Histogram
	NewHistogramVec(opts HistogramOpts, labelNames []string) Vec[Histogram]
	NewSummary(opts SummaryOpts) Summary
	NewSummaryVec(opts SummaryOpts, labelNames []string) Vec[Summary]
	Document() []MetricDescription
}

const (
	typeCounter   = "counter"
	typeGauge     = "gauge"
	typeHistogram = "histogram"
	typeSummary   = "summary"
)

type MetricDescription struct {
	Type   string   `json:"type"`
	Name   string   `json:"name"`
	Help   string   `json:"help"`
	Labels []string `json:"labels"`
}

type factory struct {
	mu      sync.Mutex
	metrics []MetricDescription
	factory promauto.Factory
}

func With(registry *prometheus.Registry) Factory {
	if registry == nil {
		registry = DefaultRegistry
	}

	return &factory{
		factory: promauto.With(registry),
	}
}

func (d *factory) NewCounter(opts CounterOpts) Counter {
	if !Enabled() {
		return noopCounter{}
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	d.metrics = append(d.metrics, MetricDescription{
		Type: typeCounter,
		Name: fullName(opts.Namespace, opts.Subsystem, opts.Name),
		Help: opts.Help,
	})
	return d.factory.NewCounter(prometheus.CounterOpts(opts))
}

func (d *factory) NewCounterVec(opts CounterOpts, labelNames []string) Vec[Counter] {
	if !Enabled() {
		return noopCounter{}
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	d.metrics = append(d.metrics, MetricDescription{
		Type:   typeCounter,
		Name:   fullName(opts.Namespace, opts.Subsystem, opts.Name),
		Help:   opts.Help,
		Labels: labelNames,
	})
	return counterVec{d.factory.NewCounterVec(prometheus.CounterOpts(opts), labelNames)}
}

func (d *factory) NewGauge(opts GaugeOpts) Gauge {
	if !Enabled() {
		return noopGauge{}
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	d.metrics = append(d.metrics, MetricDescription{
		Type: typeGauge,
		Name: fullName(opts.Namespace, opts.Subsystem, opts.Name),
		Help: opts.Help,
	})
	return d.factory.NewGauge(prometheus.GaugeOpts(opts))
}

func (d *factory) NewGaugeVec(opts GaugeOpts, labelNames []string) Vec[Gauge] {
	if !Enabled() {
		return noopGauge{}
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	d.metrics = append(d.metrics, MetricDescription{
		Type:   typeGauge,
		Name:   fullName(opts.Namespace, opts.Subsystem, opts.Name),
		Help:   opts.Help,
		Labels: labelNames,
	})
	return gaugeVec{d.factory.NewGaugeVec(prometheus.GaugeOpts(opts), labelNames)}
}

func (d *factory) NewHistogram(opts HistogramOpts) Histogram {
	if !Enabled() {
		return noopHistogram{}
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	d.metrics = append(d.metrics, MetricDescription{
		Type: typeHistogram,
		Name: fullName(opts.Namespace, opts.Subsystem, opts.Name),
		Help: opts.Help,
	})
	return d.factory.NewHistogram(prometheus.HistogramOpts(opts))
}

func (d *factory) NewHistogramVec(opts HistogramOpts, labelNames []string) Vec[Histogram] {
	if !Enabled() {
		return noopHistogram{}
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	d.metrics = append(d.metrics, MetricDescription{
		Type:   typeHistogram,
		Name:   fullName(opts.Namespace, opts.Subsystem, opts.Name),
		Help:   opts.Help,
		Labels: labelNames,
	})
	return histogramVec{d.factory.NewHistogramVec(prometheus.HistogramOpts(opts), labelNames)}
}

func (d *factory) NewSummary(opts SummaryOpts) Summary {
	if !Enabled() {
		return noopSummary{}
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	d.metrics = append(d.metrics, MetricDescription{
		Type: typeSummary,
		Name: fullName(opts.Namespace, opts.Subsystem, opts.Name),
		Help: opts.Help,
	})
	return d.factory.NewSummary(prometheus.SummaryOpts(opts))
}

func (d *factory) NewSummaryVec(opts SummaryOpts, labelNames []string) Vec[Summary] {
	if !Enabled() {
		return noopSummary{}
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	d.metrics = append(d.metrics, MetricDescription{
		Type:   typeSummary,
		Name:   fullName(opts.Namespace, opts.Subsystem, opts.Name),
		Help:   opts.Help,
		Labels: labelNames,
	})
	return summaryVec{d.factory.NewSummaryVec(prometheus.SummaryOpts(opts), labelNames)}
}

func (d *factory) Document() []MetricDescription {
	d.mu.Lock()
	defer d.mu.Unlock()
	metrics := make([]MetricDescription, len(d.metrics))
	copy(metrics, d.metrics)
	return metrics
}

func fullName(ns, subsystem, name string) string {
	out := ns
	if subsystem != "" {
		out += "_" + subsystem
	}
	return out + "_" + name
}
