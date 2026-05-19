package storage

import (
	"dos/internal/common/metrics"
)

type StorageServiceMetrics struct {
	OpSlotsInUse           metrics.Gauge
	OpSlotsAcquireDuration metrics.Histogram
	OpSlotsHoldDuration    metrics.Histogram

	UploadsSuccessDuration metrics.Histogram
	UploadsFailedDuration  metrics.Histogram

	SendsSuccessDuration metrics.Histogram
	SendsFailedDuration  metrics.Histogram

	ReplicateSuccessDuration metrics.Histogram
	ReplicateFailedDuration  metrics.Histogram

	ChunksCount      metrics.Gauge
	ChunksTotalBytes metrics.Gauge

	HeartbeatFailedTotal metrics.Counter
}
