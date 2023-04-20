package scanner

import (
	"strconv"
	"time"

	"github.com/Exca-DK/node-util/service/metrics"
)

type Recorder interface {
	RecordScan(time.Duration)
	RecordPortScan(bool)
}

func newRecorder(factory metrics.Factory) Recorder {
	r := recorder{
		scanMeter: factory.NewSummary(metrics.SummaryOpts{
			Namespace:  "scanner",
			Name:       "scan_duration",
			Help:       "amount elapsed to make whole scan",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		}),
		scanCount: factory.NewCounter(metrics.CounterOpts{
			Namespace: "scanner",
			Name:      "scan_count",
			Help:      "amount of scans done",
		}),
		portScan: factory.NewCounterVec(metrics.CounterOpts{
			Namespace: "scanner",
			Name:      "port_count",
			Help:      "amount of ports scanned",
		}, []string{"open"}),
	}

	return &r
}

type recorder struct {
	scanMeter metrics.Summary
	scanCount metrics.Counter
	portScan  metrics.Vec[metrics.Counter]
}

func (r recorder) RecordScan(duration time.Duration) {
	r.scanMeter.Observe(duration.Seconds())
	r.scanCount.Inc()
}

func (r recorder) RecordPortScan(open bool) {
	r.portScan.WithLabelValues(strconv.FormatBool(open)).Inc()
}
