package metrics

import "sync/atomic"

// Metrics collects in-memory counters.
type Metrics struct {
	Processed atomic.Int64
	Skipped   atomic.Int64
	Failed    atomic.Int64
	BytesIn   atomic.Int64
	BytesOut  atomic.Int64
	BytesSave atomic.Int64
}
