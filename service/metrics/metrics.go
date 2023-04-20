package metrics

import "sync/atomic"

var enabled atomic.Bool

func Enabled() bool {
	return enabled.Load()
}

func Enable() {
	enabled.Store(true)
}
