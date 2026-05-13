package mq

import "sync/atomic"

var mqMetrics struct {
	published uint64
	consumed  uint64
	dlq       uint64
}

type MetricsSnapshot struct {
	Published uint64
	Consumed  uint64
	DLQ       uint64
}

func ResetMetrics() {
	atomic.StoreUint64(&mqMetrics.published, 0)
	atomic.StoreUint64(&mqMetrics.consumed, 0)
	atomic.StoreUint64(&mqMetrics.dlq, 0)
}

func IncPublished() {
	atomic.AddUint64(&mqMetrics.published, 1)
}

func IncConsumed() {
	atomic.AddUint64(&mqMetrics.consumed, 1)
}

func IncDLQ() {
	atomic.AddUint64(&mqMetrics.dlq, 1)
}

func SnapshotMetrics() MetricsSnapshot {
	return MetricsSnapshot{
		Published: atomic.LoadUint64(&mqMetrics.published),
		Consumed:  atomic.LoadUint64(&mqMetrics.consumed),
		DLQ:       atomic.LoadUint64(&mqMetrics.dlq),
	}
}
