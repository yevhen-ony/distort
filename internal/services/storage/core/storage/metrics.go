package storage

import (
	"dos/internal/common/metrics"
)

type StorageMetrics struct {
	OpSlotsInUse           metrics.Gauge
	OpSlotsAcquireDuration metrics.Histogram
	OpSlotsHoldDuration    metrics.Histogram

	UploadsSuccessDuration metrics.Histogram
	UploadsFailedDuration  metrics.Histogram

	SendsSuccessDuration metrics.Histogram
	SendsFailedDuration  metrics.Histogram

	ReplicateSuccessDuration metrics.Histogram
	ReplicateFailedDuration  metrics.Histogram
}

type ChunkCatalogMetrics struct {
	ChunksCount      metrics.Gauge
	ChunksTotalBytes metrics.Gauge
}

type HeartbeatMetrics struct {
	HeartbeatFailedTotal metrics.Counter
}

func NewChunkCatalogMetrics(provider metrics.Provider) *ChunkCatalogMetrics {
	chunksCount := provider.Gauge(metrics.GaugeOpts{
		Name: "storage_catalog_chunks_count",
		Help: "Current number of chunks in the storage catalog.",
	})
	chunksTotalBytes := provider.Gauge(metrics.GaugeOpts{
		Name: "storage_catalog_chunks_total_bytes",
		Help: "Current total size in bytes of chunks tracked in the storage catalog.",
	})

	return &ChunkCatalogMetrics{
		ChunksCount:      chunksCount,
		ChunksTotalBytes: chunksTotalBytes,
	}
}

func NewHeartbeatMetrics(provider metrics.Provider) *HeartbeatMetrics {
	heartbeatFailedTotal := provider.Counter(metrics.CounterOpts{
		Name: "storage_heartbeat_failed_total",
		Help: "Total number of failed heartbeat requests from storage to master.",
	})

	return &HeartbeatMetrics{
		HeartbeatFailedTotal: heartbeatFailedTotal,
	}
}

func NewStorageMetrics(provider metrics.Provider) *StorageMetrics {
	durationBuckets := []float64{0.001, 0.01, 0.05, 0.1, 0.5, 1, 5, 10}

	opSlotsInUse := provider.Gauge(metrics.GaugeOpts{
		Name: "storage_op_slots_in_use",
		Help: "Current number of occupied heavy-operation slots in storage.",
	})
	opSlotsAcquireDuration := provider.Histogram(metrics.HistogramOpts{
		Name:    "storage_op_slots_acquire_duration_seconds",
		Help:    "Time spent acquiring a heavy-operation slot in storage.",
		Buckets: durationBuckets,
	})
	opSlotsHoldDuration := provider.Histogram(metrics.HistogramOpts{
		Name:    "storage_op_slots_hold_duration_seconds",
		Help:    "Time a heavy-operation slot remains occupied in storage.",
		Buckets: durationBuckets,
	})
	uploadsSuccessDuration := provider.Histogram(metrics.HistogramOpts{
		Name:    "storage_uploads_success_duration_seconds",
		Help:    "Duration of successful chunk upload commit operations in storage.",
		Buckets: durationBuckets,
	})
	uploadsFailedDuration := provider.Histogram(metrics.HistogramOpts{
		Name:    "storage_uploads_failed_duration_seconds",
		Help:    "Duration of failed or aborted chunk upload operations in storage.",
		Buckets: durationBuckets,
	})
	sendsSuccessDuration := provider.Histogram(metrics.HistogramOpts{
		Name:    "storage_sends_success_duration_seconds",
		Help:    "Duration of successful chunk send attempts from storage to another node.",
		Buckets: durationBuckets,
	})
	sendsFailedDuration := provider.Histogram(metrics.HistogramOpts{
		Name:    "storage_sends_failed_duration_seconds",
		Help:    "Duration of failed chunk send attempts from storage to another node.",
		Buckets: durationBuckets,
	})
	replicateSuccessDuration := provider.Histogram(metrics.HistogramOpts{
		Name:    "storage_replicate_success_duration_seconds",
		Help:    "Duration of successful replication handoff requests from storage to another node.",
		Buckets: durationBuckets,
	})
	replicateFailedDuration := provider.Histogram(metrics.HistogramOpts{
		Name:    "storage_replicate_failed_duration_seconds",
		Help:    "Duration of failed replication handoff requests from storage to another node.",
		Buckets: durationBuckets,
	})

	return &StorageMetrics{
		OpSlotsInUse:             opSlotsInUse,
		OpSlotsAcquireDuration:   opSlotsAcquireDuration,
		OpSlotsHoldDuration:      opSlotsHoldDuration,
		UploadsSuccessDuration:   uploadsSuccessDuration,
		UploadsFailedDuration:    uploadsFailedDuration,
		SendsSuccessDuration:     sendsSuccessDuration,
		SendsFailedDuration:      sendsFailedDuration,
		ReplicateSuccessDuration: replicateSuccessDuration,
		ReplicateFailedDuration:  replicateFailedDuration,
	}
}
