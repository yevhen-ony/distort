package inventory

import "dos/internal/common/metrics"

type ChunkInventoryMetrics struct {
	ChunksCount      metrics.Gauge
	ChunksTotalBytes metrics.Gauge
}

func NewChunkInventoryMetrics(provider metrics.Provider) *ChunkInventoryMetrics {
	chunksCount := provider.Gauge(metrics.GaugeOpts{
		Name: "storage_catalog_chunks_count",
		Help: "Current number of chunks in the storage catalog.",
	})
	chunksTotalBytes := provider.Gauge(metrics.GaugeOpts{
		Name: "storage_catalog_chunks_total_bytes",
		Help: "Current total size in bytes of chunks tracked in the storage catalog.",
	})

	return &ChunkInventoryMetrics{
		ChunksCount:      chunksCount,
		ChunksTotalBytes: chunksTotalBytes,
	}
}
