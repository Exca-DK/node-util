package crawl

import (
	"github.com/Exca-DK/node-util/crawler/metrics"
	base_metrics "github.com/Exca-DK/node-util/service/metrics"
)

type Recorder interface {
	RecordUselessNode()
	RecordTargetedNode()
	RecordLookup()
	RecordRecvNode()
}

func newRecorder() Recorder {
	if !metrics.Enabled() {
		return noOpRecorder()
	}
	r := crawlRecorder{
		targeted: metrics.Factory.NewCounter(base_metrics.CounterOpts{
			Namespace: "crawl",
			Subsystem: "crawler",
			Name:      "targeted_nodes",
			Help:      "amount of nodes meeting filter requirements",
		}),
		useless: metrics.Factory.NewCounter(base_metrics.CounterOpts{
			Namespace: "crawl",
			Subsystem: "crawler",
			Name:      "useless_nodes",
			Help:      "amount of nodes not meeting filter requirements",
		}),
		recv: metrics.Factory.NewCounter(base_metrics.CounterOpts{
			Namespace: "crawl",
			Subsystem: "crawler",
			Name:      "recv_nodes",
			Help:      "amount of received nodes",
		}),
		lookup: metrics.Factory.NewCounter(base_metrics.CounterOpts{
			Namespace: "crawl",
			Subsystem: "crawler",
			Name:      "lookup_iterations",
			Help:      "amount of lookup done of accepeted nodes",
		}),
	}

	return &r
}

type crawlRecorder struct {
	targeted base_metrics.Counter
	useless  base_metrics.Counter
	lookup   base_metrics.Counter
	recv     base_metrics.Counter
}

func (r crawlRecorder) RecordTargetedNode() { r.targeted.Inc() }
func (r crawlRecorder) RecordUselessNode()  { r.useless.Inc() }
func (r crawlRecorder) RecordRecvNode()     { r.recv.Inc() }
func (r crawlRecorder) RecordLookup()       { r.lookup.Inc() }

func noOpRecorder() Recorder { return nooprunnerRecorder{} }

type nooprunnerRecorder struct{}

func (r nooprunnerRecorder) RecordTargetedNode() {}
func (r nooprunnerRecorder) RecordUselessNode()  {}
func (r nooprunnerRecorder) RecordRecvNode()     {}
func (r nooprunnerRecorder) RecordLookup()       {}
