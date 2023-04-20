package runner

import (
	"time"

	"github.com/Exca-DK/node-util/crawler/metrics"
	base_metrics "github.com/Exca-DK/node-util/service/metrics"
)

type RunnerRecorder interface {
	RecordLoop()
	RecordRecvTime(d time.Duration)
	RecordOpTime(d time.Duration)
	RecordWriteTime(d time.Duration)
}

var (
	_runnerType     = []string{"runnerType"}
	_metricsCounter = metrics.Factory.NewCounterVec(base_metrics.CounterOpts{
		Namespace: "workers",
		Subsystem: "runner",
		Name:      "loops",
		Help:      "amount of iterations done by workers",
	}, _runnerType)
	_metricsSummaryRecv = metrics.Factory.NewSummaryVec(base_metrics.SummaryOpts{
		Namespace:  "workers",
		Subsystem:  "runner",
		Name:       "recv_time",
		Help:       "recv time stats across workers",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
	}, _runnerType)
	_metricsSummaryOp = metrics.Factory.NewSummaryVec(base_metrics.SummaryOpts{
		Namespace:  "workers",
		Subsystem:  "runner",
		Name:       "op_time",
		Help:       "op time stats across workers",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
	}, _runnerType)
	_metricsSummaryWrite = metrics.Factory.NewSummaryVec(base_metrics.SummaryOpts{
		Namespace:  "workers",
		Subsystem:  "runner",
		Name:       "write_time",
		Help:       "write time stats across workers",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
	}, _runnerType)
)

type noophist struct{}

func (n noophist) Observe(float64) {}

func newRecorder(runner_id string, write bool) RunnerRecorder {
	r := runnerRecorder{
		c:      _metricsCounter.WithLabelValues(runner_id),
		t_recv: _metricsSummaryRecv.WithLabelValues(runner_id),
		t_op:   _metricsSummaryOp.WithLabelValues(runner_id),
	}
	if write {
		r.t_write = _metricsSummaryWrite.WithLabelValues(runner_id)
	} else {
		r.t_write = noophist{}
	}

	return &r
}

type runnerRecorder struct {
	c       base_metrics.Counter
	t_recv  base_metrics.Summary
	t_op    base_metrics.Summary
	t_write base_metrics.Summary
}

func (r runnerRecorder) RecordLoop()                     { r.c.Inc() }
func (r runnerRecorder) RecordRecvTime(d time.Duration)  { r.t_recv.Observe(float64(d)) }
func (r runnerRecorder) RecordOpTime(d time.Duration)    { r.t_op.Observe(float64(d)) }
func (r runnerRecorder) RecordWriteTime(d time.Duration) { r.t_write.Observe(float64(d)) }

func noOpRecorder(runner_id string) RunnerRecorder {
	return nooprunnerRecorder{}
}

type nooprunnerRecorder struct{}

func (r nooprunnerRecorder) RecordLoop()                     {}
func (r nooprunnerRecorder) RecordRecvTime(d time.Duration)  {}
func (r nooprunnerRecorder) RecordOpTime(d time.Duration)    {}
func (r nooprunnerRecorder) RecordWriteTime(d time.Duration) {}
