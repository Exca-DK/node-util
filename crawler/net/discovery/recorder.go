package discovery

import (
	"strconv"
	"time"

	"github.com/Exca-DK/node-util/crawler/metrics"
	base_metrics "github.com/Exca-DK/node-util/service/metrics"
)

var labels = []string{
	"method",
	"response_code",
	"external",
}

type Recorder interface {
	RecordRpcDuration(dur time.Duration, method string, code int, ext bool)
	RecordRpcRequest(method string, code int, ext bool)
}

func newRecorder() Recorder {
	if !metrics.Enabled() {
		return noOpRecorder()
	}
	r := recorder{
		rpcHist: metrics.Factory.NewHistogramVec(base_metrics.HistogramOpts{
			Namespace: "discovery",
			Subsystem: "v4",
			Name:      "rpc_duration",
			Help:      "histogram of executed rpc methods",
		}, labels),
		rpcCounter: metrics.Factory.NewCounterVec(base_metrics.CounterOpts{
			Namespace: "discovery",
			Subsystem: "v4",
			Name:      "rpc_counter",
			Help:      "histogram of executed rpc methods",
		}, labels),
	}

	return &r
}

type recorder struct {
	rpcHist    base_metrics.Vec[base_metrics.Histogram]
	rpcCounter base_metrics.Vec[base_metrics.Counter]
}

func (r recorder) RecordRpcDuration(dur time.Duration, method string, code int, ext bool) {
	r.rpcHist.WithLabelValues(method, strconv.Itoa(code), strconv.FormatBool(ext)).Observe(float64(dur.Milliseconds()))
}

func (r recorder) RecordRpcRequest(method string, code int, ext bool) {
	r.rpcCounter.WithLabelValues(method, strconv.Itoa(code), strconv.FormatBool(ext)).Inc()
}

func noOpRecorder() Recorder { return noopRecorder{} }

type noopRecorder struct{}

func (r noopRecorder) RecordRpcDuration(dur time.Duration, method string, code int, ext bool) {}
func (r noopRecorder) RecordRpcRequest(method string, code int, ext bool)                     {}
