package broker

import (
	"github.com/Exca-DK/node-util/service/metrics"
)

type Recorder interface {
	RecordEgress(size int)
	RecordIngress(size int)
}

func NewRecorder(factory metrics.Factory) Recorder {
	r := recorder{
		egressMeter: factory.NewCounter(metrics.CounterOpts{
			Namespace: "broker",
			Name:      "egress_amount",
			Help:      "amount of data sent through broker",
		}),
		egressCount: factory.NewCounter(metrics.CounterOpts{
			Namespace: "broker",
			Name:      "egress_count",
			Help:      "how many times data has been sent through broker",
		}),
		ingressMeter: factory.NewCounter(metrics.CounterOpts{
			Namespace: "broker",
			Name:      "ingress_amount",
			Help:      "amount of data recv through broker",
		}),
		ingressCount: factory.NewCounter(metrics.CounterOpts{
			Namespace: "broker",
			Name:      "ingress_count",
			Help:      "how many times data has been recv through broker",
		}),
	}

	return &r
}

type recorder struct {
	egressMeter  metrics.Counter
	egressCount  metrics.Counter
	ingressMeter metrics.Counter
	ingressCount metrics.Counter
}

func (r recorder) RecordEgress(size int) {
	r.egressMeter.Add(float64(size))
	r.egressCount.Inc()
}

func (r recorder) RecordIngress(size int) {
	r.ingressMeter.Add(float64(size))
	r.ingressCount.Inc()
}
